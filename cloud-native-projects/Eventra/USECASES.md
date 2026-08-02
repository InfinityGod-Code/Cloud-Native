## RealTime Usecases
Realtime usecases are covered here with the real time situations and how to tackle with resilient, scalable and secured system designs.


USECASE 1 : Booking Service Under Pressure 
This is the case when the booking service gots 10,000 requests concurrently and we need server them all with ultra secure, low-latency, and consistent results.
```
                                      USER REQUESTS
                               (10,000 Concurrent Users)
                                            │
                                            │
                                 ┌────────────────────┐
                                 │   Load Balancer    │
                                 │   (NGINX / ALB)    │
                                 └─────────┬──────────┘
                                           │
                          ┌────────────────┼─────────────────┐
                          │                │                 │
                ┌─────────▼───────┐ ┌──────▼────────┐ ┌──────▼────────┐
                │ Booking API #1  │ │ Booking API #2│ │ Booking API #N│
                │  (Stateless)    │ │  (Stateless)  │ │  (Stateless)  │
                └─────────┬────────┘ └──────┬────────┘ └──────┬────────┘
                          │                 │                  │
                          └─────────────────┼──────────────────┘
                                            │
                                            ▼
                            ┌────────────────────────────────┐
                            │       Redis Cluster            │
                            │--------------------------------│
                            │ Seat Availability Cache        │
                            │ Atomic Operations (SETNX)      │
                            │ Distributed Locks              │
                            │ Lock TTL (30 sec)              │
                            └──────────────┬─────────────────┘
                                           │
                               Lock Acquired?
                           ┌─────────┴─────────┐
                           │                   │
                        YES│                   │NO
                           ▼                   ▼
               ┌────────────────────┐   ┌──────────────────┐
               │ Publish Booking    │   │ Return           │
               │ Request to Queue   │   │ Seat Already     │
               │                    │   │ Reserved         │
               └─────────┬──────────┘   └──────────────────┘
                         │
                         ▼
            ┌───────────────────────────────────────┐
            │     Kafka / RabbitMQ / Amazon SQS     │
            │                                       │
            │     Booking Request Queue             │
            └───────────────┬───────────────────────┘
                            │
          ┌─────────────────┼────────────────────┐
          │                 │                    │
 ┌────────▼───────┐ ┌────────▼───────┐ ┌─────────▼────────┐
 │ Worker Service │ │ Worker Service │ │ Worker Service   │
 │      #1        │ │      #2        │ │       #N         │
 └────────┬───────┘ └────────┬───────┘ └─────────┬────────┘
          │                  │                   │
          └──────────────────┼───────────────────┘
                             │
                             ▼
              ┌─────────────────────────────────────┐
              │      PostgreSQL / MySQL             │
              │-------------------------------------│
              │ Optimistic Locking                  │
              │                                     │
              │ UPDATE seats                        │
              │ SET status='BOOKED',                │
              │ version=version+1                   │
              │ WHERE id=?                          │
              │ AND version=?                       │
              └───────────────┬─────────────────────┘
                              │
                   Update Successful?
                    ┌─────────┴──────────┐
                    │                    │
                 YES│                    │NO
                    ▼                    ▼
      ┌─────────────────────┐   ┌────────────────────┐
      │ Booking Confirmed   │   │ Retry / Seat Sold  │
      │ Release Redis Lock  │   │ Release Lock       │
      └─────────┬───────────┘   └────────────────────┘
                │
                ▼
      ┌──────────────────────────┐
      │ Notification Service     │
      │ SMS / Email / WebSocket  │
      └──────────────────────────┘
```