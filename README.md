# bin-hot-swap

Minimal binary hot swapper program that allows to swap the currently executed program with another one.

The swapping process is triggered upon a change in the hash of the executed binary.

It also optionally allows to set the amound of times a hot swapping is allowed

## Usage
make sure that you generated a hash of your binary using ```sha256sum ./path-to/your_binary > config/hash_binary.txt```. Then execute:
- make build **or** make run directly to execute it.

you can also optionally pass -max-swaps \<int-value> to specify the amount of hot-swapping allowed before disallowing it and just continue with the last swapped binary.

You can also pass it like this ```MAX_SWAPS=\<int_value> make run```

To stop the process you can:
- make stop 
- Ctrl + C

However, once you exceed the threshold of swapping, the main process is already stopped.
In order to stop the executed swapped binary, you can call ```make stop_swapped_binary```

# Goal
I want to have the ability to swap running binary in TEEs and at the same time be able to attest the configuration of the amount of times a swapping is allowed.
It could be used in several scenarios, such as:

- hot reloading within gramine-sgx to avoid reloading a big data amount
- starting a bare VM based TEE and allow a user to inject their code and execute it after attesting the VM initial state.
- hot patching a binary without reloading the state (in case viable)

# Note
This is just for testing and not necessarily a stable product. Use it with caution.
Moreover, this is probably usefull for a subset of use-cases and not for a general use case with TEEs. It doesn't yet provides the attestation guarantee's of the newly swapped binary because it was done dynamically after the TEE initialization/creation.
Further manual steps are necessary to provide these guarantees. 
