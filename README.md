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

A Distributed Hash Table (DHT) is a distributed system that provides a lookup service similar to a hash table: key-value pairs are stored in a DHT, and any participating node can efficiently retrieve the value associated with a given key. 
[(Wikipedia: Distributed hash table)](https://en.wikipedia.org/wiki/Distributed_hash_table?oldformat=true)

<p align="center">
  <img width="600" src="https://upload.wikimedia.org/wikipedia/commons/thumb/9/98/DHT_en.svg/1000px-DHT_en.svg.png" />
</p>

## Main Challenges
- **Autonomy and decentralization:** the nodes collectively form the system without any central coordination.
- **Fault tolerance:** the system should be reliable (in some sense) even with nodes continuously joining, leaving, and failing.
- **Scalability:** the system should function efficiently even with thousands or millions of nodes.

## Paper Reference
1. Byers, J., Considine, J., & Mitzenmacher, M. (2003, February). Simple load balancing for distributed hash tables. In International Workshop on Peer-to-Peer Systems (pp. 80-87). Springer, Berlin, Heidelberg.
2. Stoica, I., Morris, R., Liben-Nowell, D., Karger, D., Kaashoek, M., Dabek, F., & Balakrishnan, H. (2003). Chord: a scalable peer-to-peer lookup protocol for internet applications. IEEE/ACM Transactions on Networking, 11(1), 17–32. https://doi.org/10.1109/tnet.2002.808407
3. Talia, Domenico; Trunfio, Paolo (December 2010). "Enabling Dynamic Querying over Distributed Hash Tables". Journal of Parallel and Distributed Computing. 70 (12): 1254–1265. doi:10.1016/j.jpdc.2010.08.012.
4. Dougherty, H. Kimm and H. Ham, "Implementation of the Distributed Hash Tables on Peer-to-peer Networks," 2008 IEEE Sarnoff Symposium, Princeton, NJ, 2008, pp. 1-5, doi: 10.1109/SARNOF.2008.4520057.
5. Baruch Awerbuch, Christian Scheideler. "Towards a scalable and robust DHT". 2006.doi:10.1145/1148109.1148163
