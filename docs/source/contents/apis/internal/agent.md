# Agent API

This API is for communication between the Seldon Scheduler and the Seldon Agent which runs next to each inference server and manages the loading and unloading of models onto the server as well as acting as a reverse proxy in the data plane for handling requests to the inference server.

## Proto Definition

```{literalinclude} ../../../../../apis/mlops/agent/agent.proto 
:language: proto
```

