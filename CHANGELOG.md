# CHANGELOG

<!--- next entry here -->

## 0.2.8
2021-12-01

### Fixes

- handle unexpected symbols (190a847ebc96a4df62049eab4cee6cb2306dc280)

## 0.2.7
2021-11-29

### Fixes

- add missing directives (5677235d6fd72fd6286406d7ea85d06ab6c61354)

## 0.2.6
2021-10-20

### Fixes

- remove omitempty from agrs for backward compatibility (968dcde83bcad7b20928057f8c0e7e2fbbd4f4d5)

## 0.2.5
2021-10-18

### Fixes

- update lexing ParseError (888e9e42a316989620ec7dbfe7e1bf4611e36603)

## 0.2.4
2021-10-08

### Fixes

- added http upstream to resolver (71c9ffd984ef1ee25fc458bfd1c73c1cfb4a8b8a)

## 0.2.3
2021-10-01

### Fixes

- add file names to combined config directives (ac8eb89f81443237d02e5b9e51d2c6601037199d)

## 0.2.2
2021-09-30

### Fixes

- update premature end of file to be a ParseError (ff8b7428284fa84fa399df031732561be97c7474)

## 0.2.1
2021-09-22

### Fixes

- updated crossplane parser to support resolver in stream upstream (bc762ce7ba9bea18f3078b59c7013f20f4a6bf8e)

## 0.2.0
2021-09-10

### Features

- Updates ParseError to export fields (64785758d836706a87fcda643fe4ae2c5e9491bf)

## 0.1.7
2021-09-10

### Fixes

- add helper for equal and stringer for directive (5ede0a29e3efc081e304f30378e07e307e1cd403)

## 0.1.4
2021-05-14

### Fixes

- add set directive to the stream>server (c15e045dc5d6d5200f76d8762bc2ebd8b4fb4f70)
- restore payload defaults, avoid side effects in performIncludes (01f904337ad277b5e363e6c9b66d511d63a54da3)

## 0.1.3
2021-05-12

### Fixes

- we should not panic in the parsing when args are missing (6e5ade38ec520e14770bdd1689494a9688d026c6)

## 0.1.2
2021-05-12

### Fixes

- you should not have pointers to slices (cdc54848b53b8903a1d7100d8abcca7f27dd0672)

## 0.1.0
2021-04-14

### Features

- fork of Arie's golang version of crossplane (bb34e99eb21010568f655462b2ec8382fb70ee6c)