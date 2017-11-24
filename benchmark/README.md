# November 2017

Fastest TCP Server Competition

## Instructions

Write a server that adheres to the below requirements. If any of the requirements are NOT met, the submission will be
disqualified from the official judging. You’re free to run the code anywhere you’d like; ie vm, docker, cloud, etc.

* Responds on port 3280.
* Take in a 10 digit number larger than 1,000,000, with support for leading 0’s. Ie. 0001000000.
* The server should shutdown if it receives “shutdown” as input.
* The server should kill any connection that sends it malformed data.
* The server should support no more than 6 connections.
* Input should be terminated with a newline sequence.
* Unique entries should be written to a file called “data.0.log” when received and this file should be created or
	cleared each time the application starts.
* Every 5 seconds, counters should be printed out to STDOUT for the number of unique numbers received, the total for
	that period, and the total for the duration of time the server has been running. The counters should then be flushed.
* Every 10 seconds, the log should rotate and increment the number in the name, all while only writing unique numbers.
	Example: data.0.log -> data.1.log -> data.2.log.

## Testing

You can use the benchmarker code in the benchmark directory to test your servers. Please feel free to open a PR for that
if there are any issues.
