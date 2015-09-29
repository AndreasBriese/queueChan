# queueCha

Go/Golang fast FIFO queue (*Posh*-*Pop*) using a channel for underlying structure. *PopChan* provides a channel with the element instead of the element itself. 

Can optionally be used as a ring-buffer by adding popped front elements back to the queues end: *PopPush* or *PopChanPush*
Ring-buffer might be shifted up or down by n steps: *Rotate(n)*

Threadsafe versions of functions for concurrent use in cororoutines: *PushTS*, *PopTS*, *PopChanTS*, *PopPushTS*, *PopChanPushTS*, *RotateTS*

See \_test.go for example.

MIT Licened, eduToolbox@BriC GmbH, Sarstedt, Andreas Briese 2015

