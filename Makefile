
swap:
	sha256sum ./binaryB/hello.go > config/hash_binary.txt

swap_back:
	sha256sum ./binaryA/hello.go > config/hash_binary.txt

build:
	go build -ldflags "-extldflags '-Wl,-z,stack-size=0x800000,-fuse-ld=gold'" -tags urfave_cli_no_docs -trimpath -v

run:
ifdef MAX_SWAPS
	./binHotSwap -max-swaps $(MAX_SWAPS) & echo $$! > go_process_pid.txt
else
	./binHotSwap & echo $$! > go_process_pid.txt
endif

stop:
	kill -TERM $(shell cat go_process_pid.txt)
#	pkill -TERM -P $(shell cat go_process_pid.txt)

clean:
	go clean
	rm -f go_process_pid.txt
	rm -f current_pid.txt

stop_swapped_binary:
	pkill -TERM -P $(shell cat current_pid.txt)
	kill -9 $(shell cat go_process_pid.txt)
	
.PHONY: swap swap_back build run clean stop_swapped_binary stop

.DEFAULT_GOAL := build

