# Pokemon API
API Sample for Nology

## Endpoints
* `/` Simply states the API is live
* `/pokemon` Lists the Pokemon the "player" has caught
* `/pokemon/id` Displays information about a particular Pokemon
* `/teams` Lists the current teams aptly named `goto` and `AST` (Abstract Syntax Tree)
* `/teams/name` Displays that specific team's roster

## Best practices I pointed out throughout the code
* Be aware of what libraries are available, discuss them with the team, and use those that follow your organization's guidelines
* Use secrets to keep sensitive information out of source files
* Implementation will vary from languauge to language, but generally ensure any opened data sources are safely closed
* Employ defensive programming, for example careful conversions
* Plan out proper development and testing environments
