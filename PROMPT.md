## Used prompts using Cursor to get speed and help to decide some technical and architectural decisions

1. Can you provide me a list of geocoding API providers for US addresses that has free tier?

2. Scaffold a folder structure for a new project called "address-validator" with the following folders for a project using golang 1.24, like 'cmd', 'internatl' and what will be necessary for a project aiming to develop using clean architecture and dependency injection.

3. Create a Dockerfile for the project and docker compose to up the application with redis.

4. Create makefile for most common commands like build, run, test, lint, etc for my project.

5. I will use two providers for geocoding. Adjust my service of geocoding to use a primary provider, called Geocoding A and a fallback provider, called Geocoding B. At first moment, I don't know what I'll use, so keep it clean so I can change it later.
Use the B provider as fallback provider, so tha primary should be called first and in case of error, the B provider should be called.

6. create a model for address validation response, because independtly of which API used, should return same validated output.

7. Create a cache service to cache the results of the geocoding, where the key of cache should be the normalized address and applied MD5. The application should query first in cache and if the info is not there, the use the external API.

8. Create unit tests for each service and special integration test for Cache, using testcontainers to intereact witha instance of Redis, instead of just mocked.

9. Add some tool to generate the swagger documentation for the API automatically. When I need to update my documentation, I just need to run a command and the documentation will be updated. If there is a go lib, use it. Else, tell me the options.

10. Add a middleware to log the request and response.

11. Add a middleware to authenticate the request. The token should be validated using the API token.

12. Add a health check endpoint to check if the service is running.

13. Add some input validation, to handle typo errors. Is there a simple NLP or other approach to handle this? If so, use it. If not, tell me the options.

14. Explain me about this Levenshtein edit distance approach to handle typo errors. Show me an example how to use it in golang.