package outbound

// now we need to keep on delivering the events to an outbound endpoint.
// we need to store the outbound endpoint address, the topic for which we need to send
// the event and a flag. if the flag is off then we no need to send events to that event for that topic
