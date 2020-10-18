# SCHRODINGER Final Project
Final Project for CSCI 6421 

Team Members
```
Xuzheng Lu (Leader)
Zetian Zheng
Xiaoyu Shen
Guangyuan Shen
```

## Topic: Distributed Hash Table

A Distributed Hash Table (DHT) is a distributed system that provides a lookup service similar to a hash table [(Wikipedia: Distributed hash table)](https://en.wikipedia.org/wiki/Distributed_hash_table?oldformat=true). The network maintains a huge file index hash table, divided and stored on each node of the network according to certain rules, with entries in the form of (key, value). Usually, the key is the hash value of the file, and the value is the IP address where the file is stored. Given the key, the value/address stored in the node can be efficiently found and returned to the query node.

<p align="center">
  <img width="600" src="https://upload.wikimedia.org/wikipedia/commons/thumb/9/98/DHT_en.svg/1000px-DHT_en.svg.png" />
</p>

**State-of-the-art DHT algorithms (for reference):**
1. **Chord** - a protocol and algorithm for a peer-to-peer distributed hash table
[(Wikipedia: Chord (peer-to-peer))](https://en.wikipedia.org/wiki/Chord_(peer-to-peer)?oldformat=true)

2. **Pastry** - an overlay network and routing network for the implementation of a distributed hash table similar to Chord.
[(Wikipedia: Pastry (DHT))](https://en.wikipedia.org/wiki/Pastry_(DHT)?oldformat=true)

3. **Tapestry** - a peer-to-peer overlay network which provides a distributed hash table, routing, and multicasting infrastructure for distributed applications.
[(Wikipedia: Tapestry (DHT))](https://en.wikipedia.org/wiki/Tapestry_(DHT)?oldformat=true)


## Why This Topic?

Distributed Hash Tables are both fault tolerant and resilient when key/value pairs are replicated. The ability to distribute data among the peers is in strong contrast to the Blockchain model in which every node has a copy of the entire ledger. The distributed hash table was chosen because it is the foundation of many applications, such as distributed file systems, peer-to-peer technology file sharing systems, cooperative web caching, multicast, anycast, domain name system, and instant messaging. Learning distributed hash tables also helps us better understand complex applications.

## Challenges
1. **Autonomy and decentralization**
- In this DHT system, the nodes should be autonomous and decentralized, which means it has no central coordination.
- For each object, the node responsible for the object should be reachable through a short routing path.
- The number of neighbors of each node should be kept reasonable.
2. **Fault tolerance**
- The DHT system should be reliable and robust in any case, and avoid the crash from high concurrency.
- In this project, we can use algorithmic controlling of the distributed system’s components to provide the desired service (Storm C, 2012).

3. **Scalability**
- The DHT should be able to handle a growing amount of nodes be added to the current system.
- The system should be able to handle the node addition and removal:
  - repartition the affected keys on the existing node;
  - reorganize the neighbor nodes;
  - to connect new nodes to DHT through a guiding mechanism.
- For example, in Apache Cassandra, an open-source NoSQL DHT system, each node can communicate with a constant amount of other nodes, which let the system to scale linearly over a huge number of nodes [(John Hammink, 2019)](https://aiven.io/blog/an-introduction-to-apache-cassandra#:~:text=This%20is%20one%20of%20the,and%20data%20centers%20go%20down).

## Metrics

We should use the following metrics to evaluate our project/product:

1. There is no central node to coordinate the relationship between each node.
2. The system composed of multiple nodes to store different hash tables.
3. Nodes can be added or withdrawn freely.After adding a new node, the position of the node can be correctly determined. When a node exits or fails, it can exit correctly without affecting other nodes to continue.
4. The main storage of the hash table is the key-value pair.

## Programming Environment

- Language: Go Language
- Platform: AWS
- Algorithms (for reference): Chord, Pastry, Tapestry

## References
1. Zave, Pamela. "Reasoning about identifier spaces: How to make chord correct." IEEE Transactions on Software Engineering 43.12 (2017): 1144-1156.
2. R. Al-Aaridhi and K. Graffi, "Sets, lists and trees: Distributed data structures on distributed hash tables," 2016 IEEE 35th International Performance Computing and Communications Conference (IPCCC), Las Vegas, NV, 2016, pp. 1-8, doi: 10.1109/PCCC.2016.7820639.
3. Talia, Domenico, and Paolo Trunfio. "Enabling dynamic querying over distributed hash tables." Journal of Parallel and Distributed Computing 70.12 (2010): 1254-1265.
4.  Storm, Christian. Specification and Analytical Evaluation of Heterogeneous Dynamic Quorum-Based Data Replication Schemes. Springer Science & Business Media, 2012.
5. Damian Gryski. (2018). Consistent Hashing: Algorithmic Tradeoffs. Medium, https://medium.com/@dgryski/consistent-hashing-algorithmic-tradeoffs-ef6b8e2fcae8
6. Farhan Ali Khan.(2018).Chord: Building a DHT (Distributed Hash Table) in Golang. Medium. https://medium.com/techlog/chord-building-a-dht-distributed-hash-table-in-golang-67c3ce17417b