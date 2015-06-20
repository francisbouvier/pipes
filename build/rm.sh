docker stop etcd1
docker stop etcd2
docker stop etcd3
docker stop swarm_agent
docker stop swarm_manager

docker rm etcd1
docker rm etcd2
docker rm etcd3
docker rm swarm_agent
docker rm swarm_manager

rm $HOME/.pipes/discovery.json
