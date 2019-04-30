# akka-cluster-operator

Kubernetes operator for applications using Akka clustering

* verbs
  * add
  * update
  * delete

* kubectl describe help:
  * show node connectivity status
  * show akka management summaries
  * read akka events into kubernetes events

* app phases
  * install app
  * roll out config changes and updates
  * mainteance, day 2 things

* add, remove, down nodes (scale)
  * shutdown oldest first so keeps moving singleton

* separate image from configuration

* autoscaling via telemetry

* remoting serialization compatible

* precheck join config, incompatible stop before try

* rebalancing shuffle, move entities coodinated as nodes come up

* drain shards off node before shutdown

* option to full stop so streaming not split between versions

* set rolling update policy

* volume mounting, writable file system

* storage service checks
