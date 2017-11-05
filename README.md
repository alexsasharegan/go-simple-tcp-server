# go-simple-tcp-server

```sh
# valid entries
echo -n "0001000000" | nc localhost 3280
echo -n "0201036000" | nc localhost 3280
echo -n "0934759801" | nc localhost 3280
echo -n "0917385761" | nc localhost 3280
# too long
echo -n "0092097873561" | nc localhost 3280
# too short
echo -n "" | nc localhost 3280
# non-numeric
echo -n "invalid input" | nc localhost 3280
```
