import logging

def track_winners(tracking_input, tracking_output, server_socket):
    server_socket.close()
    active_agencies = 0
    total_winners = 0
    while True:
        msg = tracking_input.get()
        if msg == 'S':
            #New agency started processing
            active_agencies += 1
            logging.info("Tracker: Received S. Adding new agency, total now: {}".format(active_agencies))
        elif msg == 'F':
            #An agency has finished processing
            active_agencies -= 1
            logging.info("Tracker: Received F. Removing agency, total now: {}".format(active_agencies))
        elif msg == '?':
            logging.info("Tracker: Received ?. Puting info in output queue")
            #Agency has asked for the total winners
            if active_agencies == 0:
                logging.info("Tracker: All agencies have finished, sending total winners: {}".format(total_winners))
                #If all agencies have finished send the amount of winners
                tracking_output.put({'type':'W', 'value': total_winners})
            else:
                logging.info("Tracker: Agencies still processing, sending amount: {}".format(active_agencies))
                #If there are agencies processing send the amount left
                tracking_output.put({'type':'P', 'value': active_agencies})
        elif msg == None:
            logging.info("Tracker: Received None, returning")
            return
        else:
            #If the message isnt one of the previous then it is the amount of winners that we need to add
            total_winners += msg
            logging.info("Tracker: Received amount of winners to add, total winners now: {}".format(total_winners))