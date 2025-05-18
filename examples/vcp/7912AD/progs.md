 
## tek handshake spr/sum 82
* https://w140.com/tekwiki/images/7/7b/Handshake_spring_summer_1982.pdf
* pg 5

96 is 7912
97 vert plugin
98 horiz plugin
39 likely power supply
```sh
IFC
DCL # 96
GTL # 96
V0$ # 39
MAI 252 # 96
POS -1 # 97
RIN LOW # 97
POS .5 # 98
MOD SSW;SRC EXT; T/D 1E-03; CPL DC # 98
V/D 100E-03 # 96 mistake - actually 97??
DIG SSW # 96
ATC; READ ATC #96
V1+4.5$VX$ # 39
```

