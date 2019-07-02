docker service rm zipkin

docker service create --constraint node.role==manager --replicas 1 \
-p 9411:9411 --name zipkin --network my_network \
--update-delay 10s --with-registry-auth  \
--update-parallelism 1 openzipkin/zipkin