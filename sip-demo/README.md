# simple-sip
## open two terminal
```
python sip-demo/sip-demo.py --port 5060 --username alice

python sip-demo/sip-demo.py --port 5061 --username bob --to sip:alice@127.0.0.1:5060 --text "your text"
```
