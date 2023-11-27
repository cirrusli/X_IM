cluster的启用需要结合MySQL的分区表一起，目前仍然在测试

*MySQL的分区表：解决千万级以上数据情况下离线消息同步的问题*

主要是优化寻址的吞吐量就是提高redis读吞吐量，通常一台Redis实例的读QPS可以达到5w左右， 
也就是说，对于一个100人的群来说，如果忽略消息存储所占用的时间，消息转发的吞吐量极限值也就是500

考虑采用数据分片的方式来提高吞吐量，将一个群的消息分散到多个Redis实例上，这样就可以提高消息转发的吞吐量

除了一致性Hash的分片方案，还有一致性Hash环的方案可以降低扩容导致的数据变更。而在redis cluster中，则是采用另一种方案hash slots

采用redis cluster方案需要对现有代码做修改，并且要解决数据分片情况下批量寻址问题，不过可以使用redis-go-cluster库来执行MGET操作。