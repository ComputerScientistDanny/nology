// Program: Sample API for Nology's interview process
// Programmer: Danny Betancourt
// Date uploaded to GitHub and submitted to Richard Gurney: June 7, 2022

package main

import (
	// Standard Library
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	// Provided by the services used
	sm "cloud.google.com/go/secretmanager/apiv1"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	smpb "google.golang.org/genproto/googleapis/cloud/secretmanager/v1"

	// Provided by the community
	// Best practice example #1
	// Be aware of what libraries are available,
	// discuss them with your team, and use those that follow your organization's guidelines
	"github.com/gorilla/mux"
	"github.com/rung/go-safecast"
)

func accessSecretVersion(name string) (string, error) {
	ctx := context.Background()
	client, err := sm.NewClient(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to create secretmanager client: %v", err)
	}
	defer client.Close()

	req := &smpb.AccessSecretVersionRequest{
		Name: name,
	}

	result, err := client.AccessSecretVersion(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to access secret version: %v", err)
	}

	return string(result.Payload.Data), nil
}

/******************************
* Design consideration:
* In a more distributed system,
* one could imagine the
* database connectivity being
* in its own service.
* Same for one dedicated to
* retrieving secrets.
*******************************/
func queryCollection(collection string) *mongo.Collection {
	// Best practice example #2
	// Use secrets to keep sensitive information out of source files
	username, err := accessSecretVersion("projects/700999387650/secrets/MONGOUSER/versions/1")
	if err != nil {
		log.Fatal(err)
	}

	password, err := accessSecretVersion("projects/700999387650/secrets/MONGOPASS/versions/1") // Best practice 1
	if err != nil {
		log.Fatal(err)
	}

	serverAPIOptions := options.ServerAPI(options.ServerAPIVersion1)
	clientOptions := options.Client().
		ApplyURI("mongodb+srv://" + username + ":" + password + "@cluster0.eydnarm.mongodb.net/?retryWrites=true&w=majority").
		SetServerAPIOptions(serverAPIOptions)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	// Best practice example #3
	// Implementation will vary from languauge to language, but generally ensure any opened data sources are safely closed
	defer cancel()

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	result := client.Database("pokemon").Collection(collection)

	return result
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	fmt.Fprint(w, "API launched")
}

func pokemonHandler(w http.ResponseWriter, r *http.Request) {
	var collection *mongo.Collection
	var creatures []bson.M

	if r.URL.Path != "/pokemon" {
		http.NotFound(w, r)
		return
	}

	collection = queryCollection("creatures")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		log.Fatal(err)
	}

	if err = cursor.All(ctx, &creatures); err != nil {
		log.Fatal(err)
	}

	fmt.Fprint(w, creatures)
}

func pokemonDetailHandler(w http.ResponseWriter, r *http.Request) {
	var result bson.D

	params := mux.Vars(r)
	stringID := params["id"]
	// Best practice example #4
	// Employ defensive programming, for example careful conversions
	id, err := safecast.Atoi32(stringID)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	collection := queryCollection("creatures")
	filter := bson.D{primitive.E{Key: "id", Value: id}}

	collection.FindOne(context.TODO(), filter).Decode(&result)
	fmt.Fprint(w, result)
}

func teamsHandler(w http.ResponseWriter, r *http.Request) {
	var collection *mongo.Collection
	var teams []bson.M

	if r.URL.Path != "/teams" {
		http.NotFound(w, r)
		return
	}

	collection = queryCollection("teams")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		log.Fatal(err)
	}

	if err = cursor.All(ctx, &teams); err != nil {
		log.Fatal(err)
	}

	fmt.Fprint(w, teams)
}

func teamsDetailHandler(w http.ResponseWriter, r *http.Request) {
	var result bson.D

	params := mux.Vars(r)
	name := params["name"]

	collection := queryCollection("teams")
	filter := bson.D{primitive.E{Key: "name", Value: name}}

	collection.FindOne(context.TODO(), filter).Decode(&result)
	fmt.Fprint(w, result)
}

func main() {
	r := mux.NewRouter()

	r.HandleFunc("/", indexHandler)
	r.HandleFunc("/pokemon", pokemonHandler)
	r.HandleFunc("/pokemon/{id}", pokemonDetailHandler)
	r.HandleFunc("/teams", teamsHandler)
	r.HandleFunc("/teams/{name}", teamsDetailHandler)
	http.Handle("/", r)

	// Best practice example #5
	// Plan out proper development and testing environments
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("Defaulting to port %s", port)
	}

	log.Printf("Listening on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
