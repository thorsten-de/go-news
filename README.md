# go-news 

This is a go implementation of an RSS news feed reader. It is based on chapter 11 in [Go in Action, 2nd edition](https://www.manning.com/books/go-in-action-second-edition).


## Testing the functionality

Testing is straightforward due to the workspace. We can use the local `domain` and `storage` modules directly. They are resolved from the local filesystem instead of the remote repository.

```bash
# Terminal 1
cd worker
go run ./cmd/worker

# Terminal 2
cd api
go run ./cmd/api
```

> [!IMPORTANT] 
> In the current state, these services cannot run at the same time. This is due to restrictions in the bbold store implementation that prevent opening the same file simultaneously.


## Extending the storage module with a vector store

In this example, we use
- [Qdrant](https://qdrant.tech/) as a vector store within a docker container
- Ollama as the local model provider, running locally on my machine. Alternatively, it could be run within a Docker container.
