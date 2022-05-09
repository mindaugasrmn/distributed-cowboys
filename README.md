# distributed-cowboys
Distributed cowboys


# Generate proto files
```
make proto-all
```

# init project
```
make init
```

whole logic is simple:
1. Every cowboy notifies starter what he is ready.
2. Then starter get confirmation from cowboys what they all ready, sends them time in forward for shooting.
3. The every cowboy creates cron tab task for shooting.
4. Then time comes thay all starts shoot to ech other. Till the death.
5. The last man stading is the winner.
