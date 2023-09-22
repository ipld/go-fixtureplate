```
$ fixtureplate generate \
  'dir(~5*file:1.0kB,~5*file:~102kB,2*dir{sharded}(~10*file:51kB),file:1.0MB{zero},file:10B,file:20B)'
```

Yields:

  A directory containing:
    → Approximately 5 files of 1.0 kB
    → Approximately 5 files of approximately 102 kB
    → 2 sharded directorys containing:
      → Approximately 10 files of 51 kB
    → A file of 1.0 MB containing just zeros
    → A file of 10 B
    → A file of 20 B