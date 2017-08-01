# Stream and Parse

A simple file parser. It is agnostic to type of file (HTML / XML / plain). It is also stupid, it doens't generate trees. Tell it what you are looking for, and where it should begin and end, and it will extract it.

## Benchmark

The state of the art is `lxml`.
| Language  | Time (ms) |
| ------------- | ------------- |
| Python `lxml`  | 3.8  |
| Golang `streamandparse`  | 0.8  |