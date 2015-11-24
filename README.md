# goConsulRoundRobin
A library to store healthy service endpoints from consul, round robin through them on request and refresh from consul on a timer

# TESTING

    $ export CONSUL_IP=test SERVICE_CACHE_TIMEOUT=1000
    $ go test -v -cover
