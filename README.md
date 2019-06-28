# Nano Sonar

Utility code for module and network discovery

Stub usages:

```go
import "github.com/project-nano/sonar"

	if err = endpoint.groupListener.AddService(ServiceTypeStringCore, "kcp", endpoint.fixedListenAddress, listenPort); err != nil {
		return err
	}
	if err = endpoint.groupListener.Start(); err != nil {
		return err
	}
```



Pinger usages:

```go
import "github.com/project-nano/sonar"

pinger, err := sonar.CreatePinger(endpoint.groupAddress, endpoint.groupPort, endpoint.domain)
echo, err = pinger.Query(queryTimeout)
```

