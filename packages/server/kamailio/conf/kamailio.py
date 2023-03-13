## Kamailio - equivalent of routing blocks in Python
##
## KSR - the new dynamic object exporting Kamailio functions
## Router - the old object exporting Kamailio functions
##
import KSR
## Relevant remarks:
##  * return code -255 is used to propagate the 'exit' behaviour to the
##  parent route block function. The alternative is to use the native
##  Python function sys.exit() (or exit()) -- it throws an exception that
##  is caught by Kamailio and previents the stop of the interpreter.


import KSR as KSR
import re
import uuid
import configparser
import KamailioDatabase
import phonenumbers

LOCATION = "kamailio_location"

# global variables corresponding to defined values (e.g., flags) in kamailio.cfg
FLT_ACC=1
FLT_ACCMISSED=2
FLT_ACCFAILED=3
FLT_NATS=5

FLB_NATB=6
FLB_NATSIPPING=7

# global function to instantiate a kamailio class object
# -- executed when kamailio app_python module is initialized
def mod_init():
    KSR.info("===== from Python mod init\n")
    # dumpObj(KSR)
    return kamailio()


# -- {start defining kamailio class}
class kamailio:

    kamailioDB = None
    def __init__(self):
        KSR.info('===== kamailio.__init__\n')


    # executed when kamailio child processes are initialized
    def child_init(self, rank):
        KSR.info('===== kamailio.child_init(%d)\n' % rank)
        config = configparser.ConfigParser()
        config.read('/etc/kamailio/config.ini')
        self.kamailioDB = KamailioDatabase.KamailioDatabase(config["database"]["host"], config["database"]["database"], config["database"]["user"], config["database"]["password"])

        KSR.info("Database Connected!\n")

        return 0


    def ksr_enable_tracing(self, msg):
        KSR.setflag(22)
        KSR.siptrace.sip_trace_mode("t")
        KSR.siptrace.sip_trace()
        return 1

    # SIP request routing
    # -- equivalent of request_route{}
    def ksr_request_route(self, msg):
        KSR.pv.seti("$avp(ksr_branch_count)", 1)

        # KSR.info("===== request - from kamailio python script\n")
        # KSR.info("===== method [%s] r-uri [%s]\n" % (KSR.pv.get("$rm"),KSR.pv.get("$ru")))
        if KSR.pv.get("$Rp") == 8080 and not KSR.is_WS():
            KSR.warn("SIP request received on " + KSR.pv.getw("$Rp") + "\n")
            KSR.sl.send_reply(403, "Forbidden")
            return 0

        if KSR.is_KDMQ() and KSR.pv.get("$Rp") == 5090:
            KSR.dmq.handle_message()
            return 1

        # per request initial checks
        if self.ksr_route_reqinit(msg)==-255 :
            return 1

        # NAT detection
        if self.ksr_route_natdetect(msg)==-255 :
            return 1

        # CANCEL processing
        if KSR.is_CANCEL():
            self.ksr_enable_tracing(msg)
            if KSR.tm.t_check_trans() > 0:
                self.ksr_route_relay(msg)
            return 1

        # handle requests within SIP dialogs
        if self.ksr_route_withindlg(msg) == -255 :
            return 1

        # -- only initial requests (no To tag)

        # handle retransmissions
        if KSR.tmx.t_precheck_trans()>0 :
            KSR.tm.t_check_trans()
            return 1

        if KSR.tm.t_check_trans()==0 :
            return 1



        # record routing for dialog forming requests (in case they are routed)
        # - remove preloaded route headers
        KSR.hdr.remove("Route")
        if KSR.is_method_in("IS") :
            KSR.rr.record_route()


        # account only INVITEs
        if KSR.pv.get("$rm")=="INVITE" :
            KSR.setflag(FLT_ACC) # do accounting


        #check if call is from carrier
        if  not KSR.is_WS() and KSR.permissions.allow_source_address(1) > 0:
            self.ksr_enable_tracing(msg)
            return self.ksr_route_from_carrier(msg)

        KSR.pv.sets("$avp(uuid)", KSR.hdr.gete("X-Openline-UUID"))
        #check if call is from asterisk
        if KSR.dispatcher.ds_is_from_list(0) > 0:
            self.ksr_enable_tracing(msg)
            self.ksr_route_from_asterisk(msg)
            return 1

        #INVITE forwarded from one kamailio to the other
        if KSR.pv.get("$Rp") == 5090 and KSR.registrar.registered(LOCATION) > 0:
            self.log_info("Call from Kamailio, attempting to route to local WebRTC user")
            self.ksr_enable_tracing(msg)
            return self.ksr_route_location(msg)

        #Everything after this point should be WEBRTC
        #if not KSR.is_WS():
        #    KSR.sl.sl_send_reply(403, "Request Not Allowed")
        #    return 1

        self.ksr_enable_tracing(msg)

        # check if it is an authenticated
        if KSR.is_WS():
            if self.ksr_route_webrtc_auth(msg) == -255:
                return 1
        else:
            if self.ksr_route_sip_auth(msg) == -255:
                return 1

        # handle registrations
        if self.ksr_route_registrar(msg)==-255 :
            return 1

        return self.ksr_route_from_webrtc(msg)


    def ksr_route_webrtc_auth(self, msg):
        if KSR.auth_ephemeral.autheph_check("openline.ai") < 0:
            KSR.auth.auth_challenge("openline.ai", 1)
            return -255
        #auth passed, yay!
        KSR.auth.consume_credentials()
        return 1

    def ksr_route_sip_auth(self, msg):
        if KSR.auth_db.auth_check("openline.ai", "kamailio_subscriber",1) < 0:
            KSR.auth.auth_challenge("openline.ai", 1)
            return -255
        #auth passed, yay!
        KSR.auth.consume_credentials()
        return 1

    # wrapper around tm relay function
    def ksr_route_relay(self, msg):
        # enable additional event routes for forwarded requests
        # - serial forking, RTP relaying handling, a.s.o.
       # if KSR.is_method_in("IBSU"):
       #     if KSR.tm.t_is_set("branch_route") < 0:

        if KSR.is_method_in("ISU"):
            if KSR.tm.t_is_set("onreply_route") < 0:
                KSR.pv.seti("$avp(cancelled_branches)", 0)
                KSR.tm.t_on_reply("ksr_onreply_manage")


        if KSR.tm.t_relay() < 0:
            KSR.sl.sl_reply_error()

        return -255


    # Per SIP request initial checks
    def ksr_route_reqinit(self, msg):

        if KSR.corex.has_user_agent() > 0 :
            ua = KSR.pv.gete("$ua")
            if (ua.find("friendly")!=-1 or ua.find("scanner")!=-1
                    or ua.find("sipcli")!=-1 or ua.find("sipvicious")!=-1) :
                KSR.sl.sl_send_reply(200, "Processed")
                return -255

        if KSR.maxfwd.process_maxfwd(10) < 0 :
            KSR.sl.sl_send_reply(483,"Too Many Hops")
            return -255

        if (KSR.is_OPTIONS()
                and KSR.corex.has_ruri_user() < 0) :
            KSR.sl.sl_send_reply(200,"Keepalive")
            return -255

        if KSR.sanity.sanity_check(17895, 7)<0 :
            KSR.err("Malformed SIP message from "
                    + KSR.pv.get("$si") + ":" + str(KSR.pv.get("$sp")) +"\n")
            return -255


    # Handle requests within SIP dialogs
    def ksr_route_withindlg(self, msg):
        if KSR.siputils.has_totag()<0 :
            return 1

        self.ksr_enable_tracing(msg)
        # sequential request withing a dialog should
        # take the path determined by record-routing
        if KSR.rr.loose_route()>0 :
            if self.ksr_route_dlguri(msg)==-255 :
                return -255
            if KSR.is_BYE() :
                # do accounting ...
                KSR.setflag(FLT_ACC)
                # ... even if the transaction fails
                KSR.setflag(FLT_ACCFAILED)
            elif KSR.is_ACK() :
                # ACK is forwarded statelessly
                if self.ksr_route_natmanage(msg)==-255 :
                    return -255
            elif KSR.is_NOTIFY() :
                if "SIP/2.0 100" in KSR.pv.gete("$rb"):
                    KSR.info("Filtering Trying notify\n")
                    KSR.sl.sl_send_reply(200, "Filtered")
                    return -255
                # Add Record-Route for in-dialog NOTIFY as per RFC 6665.
                KSR.rr.record_route()
            elif KSR.is_REFER() :
                # check if it is an authenticated
                if self.ksr_route_webrtc_auth(msg) == -255:
                    return 1
                self.ksr_route_update_refer(msg)
            self.ksr_route_relay(msg)
            return -255

        if KSR.is_ACK():
            if KSR.tm.t_check_trans() > 0:
                # no loose-route, but stateful ACK
                # must be an ACK after a 487
                # or e.g. 404 from upstream server
                self.ksr_route_relay(msg)
                return -255
            else:
                # ACK without matching transaction ... ignore and discard
                return -255

        KSR.sl.sl_send_reply(404, "Not here")
        return -255

    def ksr_route_update_refer(self, msg):
        dest = KSR.pv.gete("$(hdr(Refer-To){nameaddr.uri})")
        user = KSR.pv.gete("$(hdr(Refer-To){nameaddr.uri}{uri.user})")

        if re.search("^[+]?[0-9]+$", user) is not None:
            KSR.hdr.remove("Refer-To")
            KSR.hdr.append("Refer-To: <" + dest + ";user=phone>\r\n")

        return 1

    # Handle SIP registrations
    def ksr_route_registrar(self, msg):
        if not KSR.is_REGISTER():
            return 1
        KSR.registrar.unregister(LOCATION, KSR.pv.get("$fu"))
        if KSR.isflagset(FLT_NATS) :
            KSR.setbflag(FLB_NATB)
            # do SIP NAT pinging
            KSR.setbflag(FLB_NATSIPPING)

        if KSR.registrar.save(LOCATION, 0) < 0:
            KSR.sl.sl_reply_error()

        return -255


    # User location service
    def ksr_route_location(self, msg):
        orig_ruri = KSR.pv.gete("$ru")

        rc = KSR.registrar.lookup(LOCATION)
        if rc < 0:
            KSR.tm.t_newtran()
            if rc == -1 or rc == -3:
                KSR.tm.t_send_reply(404, "Not Found")
                return -255
            elif rc == -2:
                KSR.tm.t_send_reply(405, "Method Not Allowed")
                return -255
            else:
                KSR.err("Error while looking up location %d\n" % rc)
                return -255

        home_ip = KSR.pv.gete("$(xavp(ulrcd[0]=>received){uri.param,home})")
        home_uri = "sip:" + KSR.pv.gete("$(xavp(ulrcd[0]=>received){uri.param,home})")

        KSR.tm.t_newtran()

        self.log_info("Checking if " + home_uri + " is local\n")
        if not KSR.is_myself(home_uri):
            self.log_info(home_uri + " is not local routing to home\n")
            KSR.pv.sets("$ru", orig_ruri)
            KSR.pv.sets("$fsn", "internal")
            KSR.tm.t_relay_to_proto_addr("udp", home_ip, 5090)
            return -255

        # when routing via usrloc, log the missed calls also
        if KSR.is_INVITE():
            KSR.setflag(FLT_ACCMISSED)

        self.cleanup_headers(msg)

        self.ksr_route_relay(msg)
        return -255

    def ksr_route_to_carrier(self, msg, carrier):
        self.log_info("Looking up details for carrier %s\n" % (carrier))

        result = self.kamailioDB.lookup_carrier(carrier)
        if result is None:
            self.log_info("carrier not found, rejecting the call\n")
            KSR.tm.t_send_reply(401, "Carrier Not Found")
            return -255

        self.log_info("Routing call to %s\n"% (result['domain']))
        KSR.pv.sets("$ru", KSR.pv.gete("$hdr(X-Openline-Dest)"))
        KSR.pv.sets("$rd", result['domain'])
        KSR.pv.sets("$avp(auser)", result['username'])
        KSR.pv.sets("$avp(apass)", result['ha1'])
        KSR.tm.t_on_failure("ksr_failure_trunk_auth")
        self.cleanup_headers(msg)
        return self.ksr_route_relay(msg)

    def ksr_route_transfer_invite(self, msg):
        self.log_info("Routing call transfer invite\n")
        dest = KSR.pv.gete("$hdr(X-Openline-Dest)")
        user = KSR.pv.gete("$(hdr(X-Openline-Dest){uri.user})")
        source = KSR.pv.get("$(hdr(Referred-By){nameaddr.uri})")
        self.log_info("ksr_route_transfer_invite: dest=%s user=%s source=%s\n" % (dest, user, source))

        if re.search("^[+]?[0-9]+$", user) is not None:
            self.log_info("Number found, checking if PSTN is activated\n")
            self.log_info(
                "Looking up %s in database\n" % (source))
            record = self.kamailioDB.find_sipuri_mapping(source)
            if record is None:
                self.log_info("PSTN Not activated, rejecting the call\n")
                KSR.tm.t_send_reply(401, "PSTN Calling Not Allowed")
                return -255
            self.log_info("Routing call to asterisk, cli=%s carrier=%s\n" % (record['alias'], record['carrier']))
            KSR.pv.sets("$fU", record['alias'])
            return self.ksr_route_to_carrier(msg, record['carrier'])
        else:
            self.log_info("transfer: attempt local routing\n")
            KSR.pv.sets("$ru", dest)
            self.ksr_route_location(msg)
            return 1
        return 1

    def ksr_route_from_asterisk(self, msg):
        self.log_info("Routing from asterisk\n")
        if KSR.pv.get("$hdr(Referred-By)") is not None:
            return self.ksr_route_transfer_invite(msg)

        if KSR.pv.get("$hdr(X-Openline-Origin-Carrier)") is not None:
            self.log_info("From pstn flow, routing to local user\n")
            KSR.pv.sets("$ru", KSR.pv.gete("$hdr(X-Openline-Dest)"))
            self.ksr_route_location(msg)
            return 1
        elif KSR.pv.get("$hdr(X-Openline-Dest-Carrier)") is not None:
            self.log_info("From webrtc to pstn flow, routing to carrier\n")
            return self.ksr_route_to_carrier(msg, KSR.pv.gete("$hdr(X-Openline-Dest-Carrier)"))
        else:
            self.log_info("From webrtc to webrtc flow, routing to local user\n")
            KSR.pv.sets("$ru", KSR.pv.gete("$hdr(X-Openline-Dest)"))
            self.ksr_route_location(msg)
            return 1

        return 1


    # got a call from the webrtc, check if destination is pstn or webrtc and route to asterisk
    def ksr_route_from_webrtc(self, msg):

        KSR.pv.sets("$avp(uuid)", str(uuid.uuid4()))
        self.log_info("From WebRTC: Assigning call a UUID callid=%s from=%s to=%s\n" %(KSR.pv.gete("$ci"), KSR.pv.gete("$fu"), KSR.pv.gete("$tU")))

        fromUrl = KSR.pv.gete("$fu")
        KSR.tm.t_newtran()
        if KSR.pv.gete("$rU") == "echo":
            #route to echo test
            return self.ksr_route_asterisk(msg)
        elif KSR.registrar.registered(LOCATION) > 0:
            self.log_info("Destination %s is WEBRTC\n" % (KSR.pv.get("$ru")))
            KSR.hdr.append("X-Openline-Dest-Endpoint-Type: webrtc\r\n")
            return self.ksr_route_asterisk(msg)
        elif re.search("^[+]?[0-9]+$", KSR.pv.get("$rU")) is not None:
            self.log_info("Number found, checking if PSTN is activated\n")
            self.log_info(
                "Looking up %s in database\n" % (fromUrl))
            record = self.kamailioDB.find_sipuri_mapping(fromUrl)
            if record is  None:
                self.log_info("PSTN Not activated, rejecting the call\n")
                KSR.tm.t_send_reply(401, "PSTN Calling Not Allowed")
                return -255
            self.log_info("Routing call to asterisk, cli=%s carrier=%s\n" % (record['alias'], record['carrier']))
            KSR.hdr.append("X-Openline-Dest-Endpoint-Type: pstn\r\n")
            KSR.hdr.append("X-Openline-Dest-Carrier: " + record['carrier'] + "\r\n")
            KSR.hdr.append("X-Openline-CallerID: " + record['alias'] + "\r\n")

            if fromUrl != record['sipuri']:
                KSR.hdr.append("X-Openline-User: " + record['sipuri'] + "\r\n")

            destNumber = KSR.pv.get("$rU")
            newDest = self.formatInternational(record['alias'], destNumber)

            if newDest is not None:
                self.log_info("Rewriting destination number to %s" % (newDest))
                KSR.pv.sets("$rU", newDest)


            return self.ksr_route_asterisk(msg)
        else:
            self.log_info("Destination not a number nor is registered")
            KSR.tm.t_send_reply(404, "Destination Not Found")
            return -255

    def formatInternational(self, origNumber, destNumber):
        origNumberInfo = phonenumbers.parse(origNumber, None)
        if origNumberInfo is not None:
            regionCode = phonenumbers.region_code_for_number(origNumberInfo)
            if regionCode is not None:
                destNumberInfo = phonenumbers.parse(destNumber, regionCode)
                if destNumberInfo is not None:
                    destNumber = phonenumbers.format_number(destNumberInfo, phonenumbers.PhoneNumberFormat.E164)
                    return destNumber
        return None

    # got a call from the carrier, add the carrier ID and route to asterisk
    def ksr_route_from_carrier(self, msg):
        sipuri = None
        callUuid = str(uuid.uuid4())
        KSR.pv.sets("$avp(uuid)", callUuid)
        self.log_info("From Carrier: Assigning call a UUID callid=%s from=%s to=%s\n" %(KSR.pv.gete("$ci"), KSR.pv.gete("$fU"), KSR.pv.gete("$tU")))
        KSR.hdr.remove("X-Openline-UUID")

        KSR.tm.t_newtran()

        self.log_info("Looking up %s for carrier %s in database\n" % (KSR.pv.gete("$rU"), KSR.pv.gete("$avp(carrier)")))
        result = self.kamailioDB.find_e164_mapping(KSR.pv.gete("$rU"), KSR.pv.gete("$avp(carrier)"))

        if(result is None):
            KSR.tm.t_send_reply(404, "Number not assigned")
            self.log_info("No mapping found for number\n")

            return -255

        KSR.tm.t_on_branch("ksr_branch_manage")

        KSR.hdr.append("X-Openline-Origin-Carrier: " + KSR.pv.gete("$avp(carrier)") + "\r\n")
        KSR.pv.sets("$xavu("+result['sipuri']+"=>uuid)", KSR.pv.gete("$avp(uuid)"))
        KSR.pv.sets("$xavu("+result['sipuri']+"=>dest_endpoint)", "webrtc")

        KSR.pv.sets("$ru", result['sipuri'])
        KSR.hdr.remove("X-Openline-Dest-Endpoint-Type")
        KSR.hdr.remove("X-Openline-UUID")
        KSR.hdr.remove("X-Openline-Dest-User")
        self.log_info("Routing call to %s\n" % result['sipuri'])

        if self.ksr_route_asterisk(msg,True) == -255:
            return -255
        if result['phoneuri'] != "":
            KSR.tm.t_set_disable_6xx(1)
            KSR.corex.append_branch()
            KSR.pv.sets("$ru", result['phoneuri'])
            KSR.pv.sets("$xavu(" + result['phoneuri'] + "=>uuid)", KSR.pv.gete("$avp(uuid)") + "-1")
            KSR.pv.sets("$xavu(" + result['phoneuri'] + "=>dest_endpoint)", "pstn")
            KSR.pv.sets("$xavu(" + result['phoneuri'] + "=>dest_user)", result['sipuri'])

            KSR.hdr.remove("X-Openline-Dest-Endpoint-Type")
        if self.ksr_route_asterisk(msg,True) == -255:
            return -255
        KSR.pv.seti("$avp(ksr_branch_count)", 3)
        return self.ksr_route_relay(msg)
    def ksr_route_asterisk(self, msg, fork=False):
        rc = KSR.dispatcher.ds_select_dst(0, 3)
        origDest = KSR.pv.gete("$ru")

        if not fork:
            if KSR.is_WS():
                KSR.hdr.append("X-Openline-Endpoint-Type: webrtc\r\n")
            else:
                KSR.hdr.append("X-Openline-Endpoint-Type: pstn\r\n")

            KSR.hdr.append("X-Openline-UUID: " + KSR.pv.gete("$avp(uuid)") + "\r\n")
            KSR.hdr.append("X-Openline-Dest: " + origDest + "\r\n")
            if KSR.pv.gete("$rU") != "echo":
                KSR.pv.sets("$rU", "transcode")
        else:
            if KSR.is_WS():
                KSR.pv.sets("$xavu(" + origDest + "=>endpoint)", "webrtc")
            else:
                KSR.pv.sets("$xavu(" + origDest + "=>endpoint)", "pstn")
            KSR.pv.sets("$xavu(" + origDest + "=>dest)", origDest)
            if KSR.pv.gete("$rU") != "echo":
                KSR.pv.sets("$xavu(" + origDest + "=>set_ruri)", "sip:transcode@"+KSR.pv.gete("$nh(d)"))
            else:
                KSR.pv.sets("$xavu(" + origDest + "=>set_ruri)", "sip:echo@"+KSR.pv.gete("$nh(d)"))

            KSR.info("ksr_route_asterisk: Forking to %s" % origDest)
            self.print_xavp_vars(origDest)

        if rc < 0:
            KSR.tm.t_send_reply(503, "No Media Servers Available")
            return -255

        self.log_info("Routing call to asterisk (%s) ruri (%s)\n"%(KSR.pv.gete("$nh(d)"), KSR.pv.gete("$ru")))
        if not fork:
            self.ksr_route_relay(msg)

        return 1

    # Caller NAT detection
    def ksr_route_natdetect(self, msg):
        KSR.force_rport()

        if KSR.nathelper.nat_uac_test(65)>0 :
            if KSR.is_REGISTER() :
                KSR.nathelper.fix_nated_register()
                KSR.pv.sets("$avp(RECEIVED)", KSR.pv.gete("$avp(RECEIVED)") + ";home="+ KSR.pv.gete("$Ri"))
                KSR.info("Set RECEIVED to " + KSR.pv.gete("$avp(RECEIVED)") + "\n")
            elif KSR.siputils.is_first_hop()>0 :
                KSR.nathelper.set_contact_alias()

            KSR.setflag(FLT_NATS)

        else:
            KSR.pv.sets("$avp(RECEIVED)", KSR.pv.gete("$su") + ";home=" + KSR.pv.gete("$Ri"))
        return 1


    # RTPProxy control
    def ksr_route_natmanage(self, msg):
        if KSR.siputils.is_request()>0 :
            if KSR.siputils.has_totag()>0 :
                if KSR.rr.check_route_param("nat=yes")>0 :
                    KSR.setbflag(FLB_NATB)

        if (not (KSR.isflagset(FLT_NATS) or KSR.isbflagset(FLB_NATB))) :
            return 1

        #KSR.rtpproxy.rtpproxy_manage("co")

        if KSR.siputils.is_request()>0 :
            if not KSR.siputils.has_totag() :
                if KSR.tmx.t_is_branch_route()>0 :
                    KSR.rr.add_rr_param(";nat=yes")

        if KSR.siputils.is_reply()>0 :
            if KSR.isbflagset(FLB_NATB) :
                KSR.nathelper.set_contact_alias()

        return 1


    # URI update for dialog requests
    def ksr_route_dlguri(self, msg):
        if not KSR.isdsturiset() :
            KSR.nathelper.handle_ruri_alias()

        return 1


    # Routing to foreign domains
    def ksr_route_sipout(self, msg):
        if KSR.is_myself_ruri() :
            return 1

        KSR.hdr.append("P-Hint: outbound\r\n")
        self.ksr_route_relay(msg)
        return -255

    def print_xavp_vars(self, origDest):
        self.log_info("Variables for " + origDest + ":\n")
        self.log_info("X-Openline-UUID: " + KSR.pv.gete("$xavu(" + origDest + "=>uuid)") + "\r\n")
        self.log_info("X-Openline-Dest-Endpoint-Type: " + KSR.pv.gete("$xavu(" + origDest + "=>dest_endpoint)") + "\r\n")
        self.log_info("X-Openline-Dest-User: " + KSR.pv.gete("$xavu(" + origDest + "=>dest_user)") + "\r\n")
        self.log_info("X-Openline-Endpoint-Type: " + KSR.pv.gete("$xavu(" + origDest + "=>endpoint)") + "\r\n")
        self.log_info("X-Openline-Dest: " + KSR.pv.gete("$xavu(" + origDest + "=>dest)") + "\r\n")
        self.log_info("X-Openline-Set-RURI: " + KSR.pv.gete("$xavu(" + origDest + "=>set_ruri)") + "\r\n")

    # Manage outgoing branches
    # -- equivalent of branch_route[...]{}
    def ksr_branch_manage(self, msg):

        origDest = KSR.pv.gete("$ru")
        self.log_info("new branch ["+ str(KSR.pv.get("$T_branch_idx"))
                    + "] to "+ origDest  +"\n")
        self.print_xavp_vars(origDest)
        KSR.hdr.remove("X-Openline-UUID")
        KSR.hdr.remove("X-Openline-Dest-Endpoint-Type")
        KSR.hdr.remove("X-Openline-Dest-User")
        KSR.hdr.remove("X-Openline-Endpoint-Type")
        KSR.hdr.remove("X-Openline-Dest")
        if KSR.pv.gete("$xavu(" + origDest + "=>uuid)"):
            KSR.hdr.append("X-Openline-UUID: " + KSR.pv.gete("$xavu(" + origDest + "=>uuid)") + "\r\n")
        if KSR.pv.gete("$xavu(" + origDest + "=>dest_endpoint)"):
            KSR.hdr.append("X-Openline-Dest-Endpoint-Type: " + KSR.pv.gete("$xavu(" + origDest + "=>dest_endpoint)") + "\r\n")
        if KSR.pv.gete("$xavu(" + origDest + "=>dest_user)"):
            KSR.hdr.append("X-Openline-Dest-User: " + KSR.pv.gete("$xavu(" + origDest + "=>dest_user)") + "\r\n")
        if KSR.pv.gete("$xavu(" + origDest + "=>endpoint)"):
            KSR.hdr.append("X-Openline-Endpoint-Type: " + KSR.pv.gete("$xavu(" + origDest + "=>endpoint)") + "\r\n")
        if KSR.pv.gete("$xavu(" + origDest + "=>dest)"):
            KSR.hdr.append("X-Openline-Dest: " + KSR.pv.gete("$xavu(" + origDest + "=>dest)") + "\r\n")
        if KSR.pv.gete("$xavu(" + origDest + "=>set_ruri)"):
            KSR.pv.sets("$ru", KSR.pv.gete("$xavu(" + origDest + "=>set_ruri)"))

        self.log_info("Route new branch: next hop is " + KSR.pv.gete("$nh(u)") + "\n")
        self.ksr_route_natmanage(msg)
        return 1


    # Manage incoming replies
    # -- equivalent of onreply_route[...]{}
    def ksr_onreply_manage(self, msg):
        scode = KSR.pv.get("$rs")
        branchId = KSR.pv.gete("$T_branch_idx")
        self.log_info("incoming reply %d on branch %d\n"  % (scode, branchId))

        if scode>100 and scode<299 :
            self.ksr_route_natmanage(msg)
            KSR.pv.sets("$avp(answered)", "yes")
        #hack for demo
        # if scode == 183:
        #     numBranches = KSR.pv.gete("$avp(ksr_branch_count)")
        #     if numBranches > 1 and branchId != 1:
        #         self.log_info("dropping early media\n")
        #         KSR.set_drop()
        #
        #
        # if scode > 399:
        #     if KSR.pv.gete("$avp(answered)") == "yes":
        #         self.log_info("call already answered dropping error reply\n")
        #         KSR.set_drop()
        #         return 1
        #
        #     numBranches = KSR.pv.gete("$avp(ksr_branch_count)")
        #     cancelledBranches = KSR.pv.gete("$avp(cancelled_branches)")
        #     self.log_info("numBranches: %s canncelledBranches: %s\n" % (numBranches, cancelledBranches))
        #     if numBranches - 1 > cancelledBranches:
        #         KSR.set_drop()
        #         alreadyCancelled = KSR.pv.gete("$xavu(cancelled=>"+str(KSR.pv.get("$T_branch_idx"))+")")
        #         if alreadyCancelled != "yes":
        #             KSR.pv.sets("$xavu(cancelled=>"+str(KSR.pv.get("$T_branch_idx"))+")", "yes")
        #             KSR.pv.seti("$avp(cancelled_branches)", cancelledBranches + 1)
        #         else:
        #             self.log_info("branch already cancelled previously not incrementing\n")
        #         self.log_info("dropping error response \n")
        return 1

    def ksr_onsend_route(self, msg):
        return 1

    # Manage failure routing cases
    # -- equivalent of failure_route[...]{}
    def ksr_failure_trunk_auth(self, msg):
        if self.ksr_route_natmanage(msg)==-255 : return 1

        if KSR.tm.t_is_canceled()>0 :
            return 1

        if KSR.tm.t_check_status("401|407") > 0:
            KSR.info("Sending Digest Challenge with user %s\n" % (KSR.pv.gete("$avp(auser)")))
            if KSR.uac.uac_auth_mode(1) > 0:
                KSR.tm.t_relay()
        return 1


    # SIP response handling
    # -- equivalent of reply_route{}
    def ksr_reply_route(self, msg):
        KSR.dbg("response handling - python script\n")

        if KSR.sanity.sanity_check(17604, 6)<0 :
            KSR.err("Malformed SIP response from "
                    + KSR.pv.get("$si") + ":" + str(KSR.pv.get("$sp")) +"\n")
            KSR.set_drop()
            return -255

        return 1

    def ksr_xhttp_event(self, msg, evname):
        KSR.dbg("===== xhttp module triggered event:\n")
        KSR.set_reply_close()
        KSR.set_reply_no_connect()
        if KSR.pv.get("$Rp") != 8080:
            KSR.xhttp.xhttp_reply(403, "Forbidden", "" "")
            return -255

        if re.search("websocket", KSR.pv.getw("$hdr(Upgrade)").lower()) is not None and re.search("upgrade", KSR.pv.getw("$hdr(Connection)").lower()) is not None and re.search("GET", KSR.pv.getw("$rm")) is not None :
            if KSR.websocket.handle_handshake()>0 :
                return 1
            else:
                KSR.error("Websocket handshake failed\n")
        else:
            KSR.dbg("not a ws request\n")
        KSR.xhttp.xhttp_reply(200, "Ping", "text/plain", "hello world")
        return 1

    def ksr_rtimer_reload(self, msg, evname):
        KSR.dbg("reloading address table\n")
        KSR.jsonrpcs.exec('{"jsonrpc": "2.0", "method": "permissions.addressReload", "id": 1}')
        KSR.dbg("reload address result: " + KSR.pv.getw("$jsonrpl(body)") + "\n")
        KSR.dbg("reloading dispatcher table\n")
        KSR.jsonrpcs.exec('{"jsonrpc": "2.0", "method": "dispatcher.reload", "id": 2}')
        KSR.dbg("reload dispatcher result: " + KSR.pv.getw("$jsonrpl(body)") + "\n")
        return 1

    def ksr_websocket_event(self, msg, evname):
        return 1

    def ksr_dispatcher_event(self, msg, evname):
        if evname == "dispatcher:dst-down":
            KSR.err("Node down! (node=%s)\n" %(KSR.pv.gete("$ru")))
        else:
            KSR.err("Node Up! (node=%s)\n" % (KSR.pv.gete("$ru")))
        return 1

    def log_info(self, msg: str):
        KSR.info("[%s] %s" % (KSR.pv.gete("$avp(uuid)"), msg))

    def cleanup_headers(self, msg):
        KSR.hdr.remove("X-Openline-Origin-Carrier")
        KSR.hdr.remove("X-Openline-Dest")
        KSR.hdr.remove("X-Openline-Origin-Carrier")
        KSR.hdr.remove("X-Openline-Dest-Carrier")
        KSR.hdr.remove("X-Openline-Dest-Endpoint-Type")
        KSR.hdr.remove("X-Openline-Endpoint-Type")
        KSR.hdr.remove("X-Openline-CallerID")


# -- {end defining kamailio class}


# global helper function for debugging purposes
def dumpObj(obj):
    for attr in dir(obj):
        KSR.info("obj.%s = %s\n" % (attr, getattr(obj, attr)))

