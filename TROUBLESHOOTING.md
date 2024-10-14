# Troubleshooting

### Drop data from DB
Sometimes, you need to drop data from DB:

```bash
docker exec -it mongodb mongosh
> use free5gc
> db.dropDatabase()
> exit # (Or Ctrl-D)
```

### How to inspect logs?
You can see logs for each service using `docker logs` command. For example, to access the logs of the _SMF_ you can use:

```console
docker logs smf
```

### MongoDB Error Resolve

If you encounter an issue where MongoDB's Docker cannot start and the following error message is generated.

Error Message on NRF

```bash
2024-10-14T05:19:50.163862777Z [ERRO][NRF][NFM] SetLocationHeader err: RestfulAPIGetOne err: server selection error: server selection timeout, current topology: { Type: Unknown, Servers: [{ Addr: db:27017, Type: Unknown, Last error: connection() error occured during connection handshake: dial tcp: lookup db on 127.0.0.11:53: server misbehaving }, ] }
```

Error Message on Mongodb
```bash
2024-10-14T05:19:17.053+0000 I STORAGE  [initandlisten] wiredtiger_open config: create,cache_size=473M,session_max=20000,eviction=(threads_min=4,threads_max=4),config_base=false,statistics=(fast),cache_cursors=false,compatibility=(release="3.0",require_max="3.0"),log=(enabled=true,archive=true,path=journal,compressor=snappy),file_manager=(close_idle_time=100000),statistics_log=(wait=0),verbose=(recovery_progress),
2024-10-14T05:19:17.752+0000 E STORAGE  [initandlisten] WiredTiger error (-31802) [1728883157:752740][1:0x7f1099fa5580], file:WiredTiger.wt, connection: unable to read root page from file:WiredTiger.wt: WT_ERROR: non-specific WiredTiger error
2024-10-14T05:19:17.752+0000 E STORAGE  [initandlisten] WiredTiger error (0) [1728883157:752839][1:0x7f1099fa5580], file:WiredTiger.wt, connection: WiredTiger has failed to open its metadata
2024-10-14T05:19:17.752+0000 E STORAGE  [initandlisten] WiredTiger error (0) [1728883157:752847][1:0x7f1099fa5580], file:WiredTiger.wt, connection: This may be due to the database files being encrypted, being from an older version or due to corruption on disk
2024-10-14T05:19:17.752+0000 E STORAGE  [initandlisten] WiredTiger error (0) [1728883157:752851][1:0x7f1099fa5580], file:WiredTiger.wt, connection: You should confirm that you have opened the database with the correct options including all encryption and compression options
2024-10-14T05:19:17.754+0000 E -        [initandlisten] Assertion: 28595:-31802: WT_ERROR: non-specific WiredTiger error src/mongo/db/storage/wiredtiger/wiredtiger_kv_engine.cpp 421
2024-10-14T05:19:17.770+0000 I STORAGE  [initandlisten] exception in initAndListen: Location28595: -31802: WT_ERROR: non-specific WiredTiger error, terminating
2024-10-14T05:19:17.771+0000 I NETWORK  [initandlisten] shutdown: going to close listening sockets...
2024-10-14T05:19:17.771+0000 I NETWORK  [initandlisten] removing socket file: /tmp/mongodb-27017.sock
2024-10-14T05:19:17.771+0000 I CONTROL  [initandlisten] now exiting
2024-10-14T05:19:17.771+0000 I CONTROL  [initandlisten] shutting down with code:100
```

#### Here is the solution.
```bash
# Remove the container first.
$ sudo docker compose rm

# List the free5GC database volumes, you should find free5gc-compose_dbdata.
$ sudo docker volume ls | grep dbdata

# Remove the old DB volume using following command.
$ sudo docker volume rm fre5gc-compose_dbdata

# And then, you can build docker container again.
$ sudo docker compose up
```