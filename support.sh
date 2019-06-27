docker service rm rabbitmq
docker build -t linhnh123/rabbitmq support/rabbitmq/
docker service create --name=rabbitmq --replicas=1 --network=my_network -p 1883:1883 -p 5672:5672 -p 15672:15672 linhnh123/rabbitmq