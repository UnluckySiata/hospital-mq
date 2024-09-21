# Health Clinic Simulation
An implementation of 3 types of users meant to simulate work of a clinic,
done using rabbitmq with load balancing. A simple diagram of defined
queues and exchanges can be found ![here](https://github.com/UnluckySiata/hospital-mq/blob/main/hospital-mq.svg)

Defined actors:
- doctor can request a health examination of a predefined type
- technician can perform examination of given types upon request
- admin monitors the whole traffic and can send info messages to everyone


## Running
Defined commands to launch a cli:
```bash
go run ./cmd/doctor -n [name]
go run ./cmd/technician [-h -k -e for different examination types, can be more than one]
go run ./cmd/admin
```

This mini-project was developed as a part of the distributed systems course at AGH UST.

