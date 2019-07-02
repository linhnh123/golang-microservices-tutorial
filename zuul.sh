docker service rm edge-server

docker service create --replicas 1 --name edge-server -p 8765:8765 \
 --network my_network --update-delay 10s --with-registry-auth \
 --update-parallelism 1 eriklupander/edge-server