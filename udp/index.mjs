// use netcat to test
// *********************
// > nc -u 127.0.0.1 5500
// hi
// hello
// *********************

import { info, log } from 'console';
import dgram from 'dgram'

const socket = dgram.createSocket("udp4");

socket.bind(5500,"127.0.0.1")

socket.on("message",(msg, info) => {
    log(`My server got a daragram ${msg}, from ${info.address}:${info.port}`)
})