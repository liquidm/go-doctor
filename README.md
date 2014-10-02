go-doctor
=========

Library to make profiling easier. Go-doctor lets you to dump cpu profiling samples and heap profile into files, so you can use them in go tool pprof. It also outputs some useful profiling informations.

Usage
=========

Import go-doctor:

```go
import(
  "github.com/liquidm/go-doctor"
)
```

Initialize it at the begging of the program:

```
doctor.StartWithFlags()
```

Configure it using following cmd line flags:

```
  -doctor=false: enable doctor
  -cpu="prof.cprof": write cpu profile to this file
  -mem="prof.mprof": write mem profile to this file
  -statsgc=true: show GC stats
  -statsgoroutine=true: show goroutine stats
  -statsmem=true: show mem stats
```


From now on, every time you press C-c go-doctor will dump some statistics and create profiling files. Hit C-c twice to exit your program.
