# MP2-hangy6-tian23 Distributed Group Membership

## Description
A Go distributed group membership with a ring topology. The failure detection is achieved by monitoring 4 successors in the ring and check the ack from the successors pinged. We choose 4 successors to monitor since three simultaneous daemon fail can happen and we want to ensure the completeness of the failure detection. Introducer will introduce a joining node to all other members in the ring system. This project is implemented by hangy6(Hang Yu) and tian23(Tian Luan).


## Installation

You can clone this project to the machines you need to grep log from using following command:

```
ssh: git clone git@gitlab.engr.illinois.edu:hangy6/mp2-hangy6-tian23.git
```
```
https: git clone https://gitlab.engr.illinois.edu/hangy6/mp2-hangy6-tian23.git
``` 

## Build

To build the ring topology, create a file named introducer on the machine that will be the introducer:

```
touch ~/introducer
```

or

```
touch /home/[NETID]/introducer
```

Please modify the introducer host to your actual introducer host in src/daemon/request.go, function Nodejoin() before you run the who program. Then under src/daemon, starter the introducer by:

```
go run .
```

Now the ring is empty and the introducer will always be there to add any node to the ring, we first introduce ourself to the ring and in the interactive shell, we run:

```
join
```

Now, the first node has joined which is a ring with one node. For other machines which are common node, simply start the daemon with：

```
go run .
```

## Usage

To join in the ring, use command:

```
join
```

To leave the ring (not a failure or killed), run:
```
leave
```

To list the current membership list, on any node run:
```
list_mem
```

To list the current monitor list on current node, run:
```
list_mon
```

To get the identity of the current node, run:
```
list_self
```

## Support
If you have any questions, please contact tian23@illinois.edu or hangy6@illinois.edu

## Authors 
Tian Luan & Hang Yu

