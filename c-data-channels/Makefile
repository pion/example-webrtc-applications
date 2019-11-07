c-data-channels: webrtc.so data-channels.c
	gcc -o $@ data-channels.c ./webrtc.so

webrtc.so: webrtc.go bridge.go 
	go build -o $@ -buildmode=c-shared $^

clean:
	rm webrtc.so webrtc.h c-data-channels