import unittest
import KSR as KSR
import TestKamailioDatabase as TestKamailioDatabase
import sys
sys.path.insert(0, "../conf/")
import kamailio as kamailio
import ksr_utils as ksr_utils


class MyTestCase(unittest.TestCase):

    def setUp(self) -> None:
        ksr_utils.ksr_utils_init(KSR._mock_data)

    def test_register(self):
        ksr_utils.pvar_set("$rm", "REGISTER")
        ksr_utils.pvar_set("$fu", "sip:test@openline.ai")
        ksr_utils.pvar_set("$ct", "<sip:10.0.0.1:9999>;expires=6000")
        ksr_utils.pvar_set("$Rp", 8080)
        ksr_utils.pvar_set("$Ri", "1.2.3.4")

        #indicate the call is not from a carrier ip address
        KSR._mock_data["permissions"]["allow_source_address"] = 0
        #indicate the call is not coming from an asterisk ip
        KSR._mock_data["dispatcher"]["ds_is_from_list"] = 0

        k = kamailio.kamailio()
        k.kamailioDB = TestKamailioDatabase.TestKamailioDatabase()
        k.ksr_request_route(None)
        print(ksr_utils.registrations["kamailio_location"][ksr_utils.pvar_get("$fu")])
        print(ksr_utils.pvar_get("$ct"))
        self.assertEqual(ksr_utils.registrations["kamailio_location"][ksr_utils.pvar_get("$fu")], ksr_utils.pvar_get("$ct"))  # add assertion here
        self.assertEqual(ksr_utils.pvar_get("$(avp(RECEIVED){uri.param,home})"), "1.2.3.4")


    def test_INVITE_from_webrtc_to_webrtc(self):
        ksr_utils.pvar_set("$rm", "INVITE")
        ksr_utils.pvar_set("$ru", "sip:test@openline.ai")
        ksr_utils.pvar_set("$fu", "sip:AgentSmith@agent.openline.ai")
        ksr_utils.pvar_set("$ct", "<sip:10.0.0.2:8080>")
        ksr_utils.pvar_set("$Rp", 8080)
        ksr_relay_called = False


        def my_relay():
            nonlocal ksr_relay_called
            print("Inside t_relay()\n")
            self.assertEqual(ksr_utils.pvar_get("$nh(u)"), "sip:dispatcher_group_0")
            ksr_relay_called = True
            return 1

        KSR._mock_data["tm"]["t_relay"] = my_relay
        #simulate a registration in the location db
        ksr_utils.registrations["kamailio_location"][ksr_utils.pvar_get("$ru")] = "<sip:10.0.0.1:9999>;expires=6000"

        #indicate the call is not from a carrier ip address
        KSR._mock_data["permissions"]["allow_source_address"] = 0
        #indicate the call is not coming from an asterisk ip
        KSR._mock_data["dispatcher"]["ds_is_from_list"] = 0

        k = kamailio.kamailio()
        k.kamailioDB = TestKamailioDatabase.TestKamailioDatabase()

        k.ksr_request_route(None)

        self.assertEqual(ksr_utils.pvar_get("$ru"), "sip:transcode@openline.ai", "RURI not set to expected destination")  # add assertion here
        self.assertEqual(ksr_utils.pvar_get("$nh(u)"),"sip:dispatcher_group_0", "Dispatcher not invoked to right dispatcher group")
        self.assertEqual(ksr_utils.pvar_get("$hdr(X-Openline-Endpoint-Type)"),"webrtc", "Endpoint header not correctly set")
        self.assertEqual(ksr_utils.pvar_get("$hdr(X-Openline-Dest-Endpoint-Type)"),"webrtc", "Dest Endpoint Header not set")

        self.assertEqual(ksr_utils.pvar_get("$hdr(X-Openline-Dest)"), "sip:test@openline.ai", "X-Openline-Dest not set!")

        self.assertIsNotNone(ksr_utils.pvar_get("$hdr(X-Openline-UUID)"), "Missing UUID Header")
        self.assertTrue(ksr_relay_called, "Call was not routed!")

    def test_INVITE_from_webrtc_to_pstn(self):
        ksr_utils.pvar_set("$rm", "INVITE")
        ksr_utils.pvar_set("$ru", "sip:+44075755588858@openline.ai")
        ksr_utils.pvar_set("$fu", "sip:AgentSmith@agent.openline.ai")
        ksr_utils.pvar_set("$ct", "<sip:10.0.0.2:8080>")
        ksr_utils.pvar_set("$Rp", 8080)
        ksr_relay_called = False


        def my_relay():
            nonlocal ksr_relay_called
            print("Inside t_relay()\n")
            self.assertEqual(ksr_utils.pvar_get("$nh(u)"), "sip:dispatcher_group_0")
            ksr_relay_called = True
            return 1


        KSR._mock_data["tm"]["t_relay"] = my_relay
        #do not register in location db

        #indicate the call is not from a carrier ip address
        KSR._mock_data["permissions"]["allow_source_address"] = 0
        #indicate the call is not coming from an asterisk ip
        KSR._mock_data["dispatcher"]["ds_is_from_list"] = 0

        k = kamailio.kamailio()
        k.kamailioDB = TestKamailioDatabase.TestKamailioDatabase()

        def mock_sipuri_mapping(sipuri:str):
            self.assertEqual(sipuri,"sip:AgentSmith@agent.openline.ai", "Incorrect key lookup in database")
            return {"e164": '+328080970',
                    "alias":'+322800000',
                    "carrier": 'test_carrier'
                    }
        k.kamailioDB._mock['find_sipuri_mapping'] = mock_sipuri_mapping


        k.ksr_request_route(None)

        self.assertEqual(ksr_utils.pvar_get("$ru"), "sip:transcode@openline.ai", "RURI not set to expected destination")  # add assertion here
        self.assertEqual(ksr_utils.pvar_get("$nh(u)"),"sip:dispatcher_group_0", "Dispatcher not invoked to right dispatcher group")
        self.assertEqual(ksr_utils.pvar_get("$hdr(X-Openline-Endpoint-Type)"),"webrtc", "Endpoint Header not set")
        self.assertEqual(ksr_utils.pvar_get("$hdr(X-Openline-Dest-Endpoint-Type)"),"pstn", "Dest Endpoint Header not set")


        self.assertIsNotNone(ksr_utils.pvar_get("$hdr(X-Openline-UUID)"), "Missing UUID Header")
        self.assertEqual(ksr_utils.pvar_get("$hdr(X-Openline-CallerID)"), "+322800000", "CallerID not SET!")
        self.assertEqual(ksr_utils.pvar_get("$hdr(X-Openline-Dest-Carrier)"), "test_carrier", "Carrier not SET!")
        self.assertEqual(ksr_utils.pvar_get("$hdr(X-Openline-Dest)"), "sip:+44075755588858@openline.ai", "X-Openline-Dest not set!")
        self.assertTrue(ksr_relay_called, "Call was not routed!")

    def test_INVITE_from_pstn_to_webrtc(self):
        ksr_utils.pvar_set("$rm", "INVITE")
        ksr_utils.pvar_set("$ru", "sip:+328080970@openline.ai")
        ksr_utils.pvar_set("$fu", "sip:+44075755588858@agent.openline.ai")
        ksr_utils.pvar_set("$ct", "<sip:10.0.0.2:8080>")
        ksr_utils.pvar_set("$Rp", 5060) #pstn is SIP, not WS
        ksr_relay_called = False
        ksr_branch_called = False
        def allow_source_address(mode: int):
            ksr_utils.pvar_set("$avp(carrier)", "test_carrier")
            return 1

        def my_t_on_branch():
            nonlocal ksr_branch_called
            ksr_branch_called = True
            return 1
        def my_relay():
            nonlocal ksr_relay_called
            nonlocal ksr_branch_called
            print("Inside t_relay()\n")
            self.assertEqual(ksr_utils.pvar_get("$nh(u)"), "sip:dispatcher_group_0")
            ksr_relay_called = True
            k.ksr_branch_manage(None)
            return 1

        KSR._mock_data["tm"]["t_relay"] = my_relay
        #do not register in location db

        #indicate the call is from a carrier ip address
        KSR._mock_data["permissions"]["allow_source_address"] = allow_source_address
        #indicate the call is not coming from an asterisk ip
        KSR._mock_data["dispatcher"]["ds_is_from_list"] = 0

        k = kamailio.kamailio()
        k.kamailioDB = TestKamailioDatabase.TestKamailioDatabase()
        def mock_e164_mapping(e164:str, carrier:str):
            self.assertEqual(e164,"+328080970", "Incorrect e164 key lookup in database")
            self.assertEqual(carrier,"test_carrier", "Incorrect carrier key lookup in database")

            return {"sipuri": 'sip:AgentSmith@agent.openline.ai',
                     "phoneuri": '',
                    }
        k.kamailioDB._mock['find_e164_mapping'] = mock_e164_mapping


        k.ksr_request_route(None)

        self.assertEqual(ksr_utils.pvar_get("$ru"), "sip:transcode@dispatcher_group_0", "RURI not set to expected destination")  # add assertion here
        self.assertEqual(ksr_utils.pvar_get("$nh(u)"),"sip:dispatcher_group_0", "Dispatcher not invoked to right dispatcher group")
        self.assertEqual(ksr_utils.pvar_get("$hdr(X-Openline-Endpoint-Type)"),"pstn", "Endpoint Header not set")

        self.assertIsNotNone(ksr_utils.pvar_get("$hdr(X-Openline-UUID)"), "Missing UUID Header")
        self.assertEqual(ksr_utils.pvar_get("$hdr(X-Openline-Origin-Carrier)"), "test_carrier", "Carrier not SET!")
        self.assertEqual(ksr_utils.pvar_get("$hdr(X-Openline-Dest)"), "sip:AgentSmith@agent.openline.ai", "X-Openline-Dest not set!")
        self.assertTrue(ksr_relay_called, "Call was not routed!")

    def test_REFER_from_webrtc(self):
        ksr_utils.hdr_append("Refer-To: <sip:torrey@openline.ai>\r\n")
        result = ksr_utils.pvar_get("$(hdr(Refer-To){nameaddr.uri}{uri.user})")
        self.assertEqual("torrey", result)

    def test_INVITE_from_asterisk_transfer_to_pstn(self):
        ksr_utils.pvar_set("$rm", "INVITE")
        ksr_utils.pvar_set("$ru", "sip:127.0.0.1")
        ksr_utils.pvar_set("$fu", "sip:transcode@agent.openline.ai")
        ksr_utils.pvar_set("$ct", "<sip:10.0.0.2:8080>")
        ksr_utils.pvar_set("$Rp", 5060)
        ksr_utils.hdr_append("X-Openline-UUID: my uuid\r\n")
        ksr_utils.hdr_append("X-Openline-Dest: sip:+44075755588858@agent.openline.ai\r\n")
        ksr_utils.hdr_append("Referred-By: <sip:AgentSmith@agent.openline.ai>\r\n")

        ksr_relay_called = False
        ksr_on_failure_called = False


        def my_relay():
            nonlocal ksr_relay_called
            print("Inside t_relay()\n")
            ksr_relay_called = True
            return 1

        def on_failure(handler: str):
            nonlocal ksr_on_failure_called
            print("Inside t_on_failure()\n")
            ksr_on_failure_called = True
            return 1


        KSR._mock_data["tm"]["t_relay"] = my_relay
        KSR._mock_data["tm"]["t_on_failure"] = on_failure

        #do not register in location db

        #indicate the call is not from a carrier ip address
        KSR._mock_data["permissions"]["allow_source_address"] = 0
        #indicate the call is coming from an asterisk ip
        KSR._mock_data["dispatcher"]["ds_is_from_list"] = 1

        k = kamailio.kamailio()
        k.kamailioDB = TestKamailioDatabase.TestKamailioDatabase()

        def mock_lookup_carrier(carrier:str):
            self.assertEqual(carrier,"test_carrier", "Incorrect carrier key lookup in database")

            return {"username": "my_username",
                        "ha1": "my_hashed_password",
                        "domain": "carrier.domain"}
        k.kamailioDB._mock['lookup_carrier'] = mock_lookup_carrier
        def mock_sipuri_mapping(sipuri:str):
            self.assertEqual(sipuri,"sip:AgentSmith@agent.openline.ai", "Incorrect key lookup in database")
            return {"e164": '+328080970',
                    "alias": '+322800000',
                    "carrier": 'test_carrier'
                    }
        k.kamailioDB._mock['find_sipuri_mapping'] = mock_sipuri_mapping

        k.ksr_request_route(None)

        self.assertEqual(ksr_utils.pvar_get("$ru"), "sip:+44075755588858@carrier.domain", "RURI not set to expected destination")  # add assertion here
        self.assertTrue(ksr_relay_called, "Call was not routed!")
        self.assertEqual("+322800000", ksr_utils.pvar_get("$fU"))
        #ensure the digest auth handler was armed
        self.assertTrue(ksr_on_failure_called, "Call was not routed!")
        self.assertEqual("my_username", ksr_utils.pvar_get("$avp(auser)"))
        self.assertEqual("my_hashed_password", ksr_utils.pvar_get("$avp(apass)"))

    def test_INVITE_from_asterisk_transfer_to_webrtc(self):
        ksr_utils.pvar_set("$rm", "INVITE")
        ksr_utils.pvar_set("$ru", "sip:127.0.0.1")
        ksr_utils.pvar_set("$fu", "sip:transcode@agent.openline.ai")
        ksr_utils.pvar_set("$ct", "<sip:10.0.0.2:8080>")
        ksr_utils.pvar_set("$Rp", 5060)
        ksr_utils.hdr_append("X-Openline-UUID: my uuid\r\n")
        ksr_utils.hdr_append("X-Openline-Dest: sip:dev@agent.openline.ai\r\n")
        ksr_utils.hdr_append("Referred-By: <sip:AgentSmith@agent.openline.ai>\r\n")
        ksr_utils.registrations["kamailio_location"]["sip:dev@agent.openline.ai"] = "sip:10.0.0.1:9999"

        ksr_relay_called = False
        ksr_on_failure_called = False


        def my_relay():
            nonlocal ksr_relay_called
            print("Inside t_relay()\n")
            ksr_relay_called = True
            return 1

        def on_failure(handler: str):
            nonlocal ksr_on_failure_called
            print("Inside t_on_failure()\n")
            ksr_on_failure_called = True
            return 1


        KSR._mock_data["tm"]["t_relay"] = my_relay
        KSR._mock_data["tm"]["t_on_failure"] = on_failure

        #do not register in location db

        #indicate the call is not from a carrier ip address
        KSR._mock_data["permissions"]["allow_source_address"] = 0
        #indicate the call is coming from an asterisk ip
        KSR._mock_data["dispatcher"]["ds_is_from_list"] = 1

        k = kamailio.kamailio()
        k.kamailioDB = TestKamailioDatabase.TestKamailioDatabase()


        k.ksr_request_route(None)

        self.assertEqual(ksr_utils.pvar_get("$ru"), "sip:10.0.0.1:9999", "RURI not set to expected destination")  # add assertion here
        self.assertTrue(ksr_relay_called, "Call was not routed!")



    def test_INVITE_from_asterisk_to_pstn(self):
        ksr_utils.pvar_set("$rm", "INVITE")
        ksr_utils.pvar_set("$ru", "sip:127.0.0.1")
        ksr_utils.pvar_set("$fu", "sip:+328080970@agent.openline.ai")
        ksr_utils.pvar_set("$ct", "<sip:10.0.0.2:8080>")
        ksr_utils.pvar_set("$Rp", 5060)
        ksr_utils.hdr_append("X-Openline-UUID: my uuid\r\n")
        ksr_utils.hdr_append("X-Openline-Dest: sip:+44075755588858@agent.openline.ai\r\n")
        ksr_utils.hdr_append("X-Openline-Dest-Carrier: test_carrier\r\n")

        ksr_relay_called = False
        ksr_on_failure_called = False


        def my_relay():
            nonlocal ksr_relay_called
            print("Inside t_relay()\n")
            ksr_relay_called = True
            return 1

        def on_failure(handler: str):
            nonlocal ksr_on_failure_called
            print("Inside t_on_failure()\n")
            ksr_on_failure_called = True
            return 1


        KSR._mock_data["tm"]["t_relay"] = my_relay
        KSR._mock_data["tm"]["t_on_failure"] = on_failure

        #do not register in location db

        #indicate the call is not from a carrier ip address
        KSR._mock_data["permissions"]["allow_source_address"] = 0
        #indicate the call is coming from an asterisk ip
        KSR._mock_data["dispatcher"]["ds_is_from_list"] = 1

        k = kamailio.kamailio()
        k.kamailioDB = TestKamailioDatabase.TestKamailioDatabase()

        def mock_lookup_carrier(carrier:str):
            self.assertEqual(carrier,"test_carrier", "Incorrect carrier key lookup in database")

            return {"username": "my_username",
                        "ha1": "my_hashed_password",
                        "domain": "carrier.domain"}
        k.kamailioDB._mock['lookup_carrier'] = mock_lookup_carrier

        k.ksr_request_route(None)

        self.assertEqual(ksr_utils.pvar_get("$ru"), "sip:+44075755588858@carrier.domain", "RURI not set to expected destination")  # add assertion here
        self.assertTrue(ksr_relay_called, "Call was not routed!")
        #ensure the digest auth handler was armed
        self.assertTrue(ksr_on_failure_called, "Call was not routed!")
        self.assertEqual("my_username", ksr_utils.pvar_get("$avp(auser)"))
        self.assertEqual("my_hashed_password", ksr_utils.pvar_get("$avp(apass)"))

    def test_INVITE_from_asterisk_to_webrtc(self):
        ksr_utils.pvar_set("$rm", "INVITE")
        ksr_utils.pvar_set("$ru", "sip:127.0.0.1")
        ksr_utils.pvar_set("$fu", "sip:+44075755588858@agent.openline.ai")
        ksr_utils.pvar_set("$ct", "<sip:10.0.0.2:8080>")
        ksr_utils.pvar_set("$Rp", 5060)
        ksr_utils.hdr_append("X-Openline-UUID: my uuid\r\n")
        ksr_utils.hdr_append("X-Openline-Dest: sip:AgentSmit@agent.openline.ai\r\n")
        ksr_utils.hdr_append("X-Openline-Origin-Carrier: test_carrier\r\n")
        ksr_utils.registrations["kamailio_location"]["sip:AgentSmit@agent.openline.ai"] = "sip:10.0.0.1:9999"

        ksr_relay_called = False

        def my_relay():
            nonlocal ksr_relay_called
            print("Inside t_relay()\n")
            ksr_relay_called = True
            return 1

        KSR._mock_data["tm"]["t_relay"] = my_relay
        # do not register in location db

        # indicate the call is not from a carrier ip address
        KSR._mock_data["permissions"]["allow_source_address"] = 0
        # indicate the call is coming from an asterisk ip
        KSR._mock_data["dispatcher"]["ds_is_from_list"] = 1

        k = kamailio.kamailio()
        k.kamailioDB = TestKamailioDatabase.TestKamailioDatabase()


        k.ksr_request_route(None)

        self.assertEqual(ksr_utils.pvar_get("$ru"), "sip:10.0.0.1:9999",
                         "RURI not set to expected destination")  # add assertion here
        self.assertTrue(ksr_relay_called, "Call was not routed!")

    def test_INVITE_from_kamailio_to_webrtc(self):
        ksr_utils.pvar_set("$rm", "INVITE")
        ksr_utils.pvar_set("$ru", "sip:AgentSmit@agent.openline.ai")
        ksr_utils.pvar_set("$fu", "sip:+44075755588858@agent.openline.ai")
        ksr_utils.pvar_set("$ct", "<sip:10.0.0.2:8080>")
        ksr_utils.pvar_set("$Rp", 5090)
        ksr_utils.hdr_append("X-Openline-UUID: my uuid\r\n")
       # ksr_utils.hdr_append("X-Openline-Dest: sip:AgentSmit@agent.openline.ai\r\n")
       # ksr_utils.hdr_append("X-Openline-Origin-Carrier: test_carrier\r\n")
        ksr_utils.registrations["kamailio_location"]["sip:AgentSmit@agent.openline.ai"] = "sip:10.0.0.1:9999"

        ksr_relay_called = False

        def my_relay():
            nonlocal ksr_relay_called
            print("Inside t_relay()\n")
            ksr_relay_called = True
            return 1

        KSR._mock_data["tm"]["t_relay"] = my_relay
        # do not register in location db

        # indicate the call is not from a carrier ip address
        KSR._mock_data["permissions"]["allow_source_address"] = 0
        # indicate the call is not coming from an asterisk ip
        KSR._mock_data["dispatcher"]["ds_is_from_list"] = 0

        k = kamailio.kamailio()
        k.kamailioDB = TestKamailioDatabase.TestKamailioDatabase()


        k.ksr_request_route(None)

        self.assertEqual(ksr_utils.pvar_get("$ru"), "sip:10.0.0.1:9999",
                         "RURI not set to expected destination")  # add assertion here
        self.assertTrue(ksr_relay_called, "Call was not routed!")

    def test_INVITE_from_asterisk_to_webrtc_non_local(self):
        ksr_utils.pvar_set("$rm", "INVITE")
        ksr_utils.pvar_set("$ru", "sip:127.0.0.1")
        ksr_utils.pvar_set("$fu", "sip:+44075755588858@agent.openline.ai")
        ksr_utils.pvar_set("$ct", "<sip:10.0.0.2:8080>")
        ksr_utils.pvar_set("$Rp", 5060)
        ksr_utils.hdr_append("X-Openline-UUID: my uuid\r\n")
        ksr_utils.hdr_append("X-Openline-Dest: sip:AgentSmit@agent.openline.ai\r\n")
        ksr_utils.hdr_append("X-Openline-Origin-Carrier: test_carrier\r\n")
        ksr_utils.registrations["kamailio_location"]["sip:AgentSmit@agent.openline.ai"] = "sip:10.0.0.1:9999"

        ksr_relay_called = False
        ksr_relay_destination = ""
        ksr_relay_port = 0


        def t_relay_to_proto_addr(proto: str, ip: str, port: int) -> int:
            nonlocal ksr_relay_called
            nonlocal ksr_relay_destination
            nonlocal ksr_relay_port
            print("Inside t_relay_to_proto_addr()\n")
            ksr_relay_called = True
            ksr_relay_destination = ip
            ksr_relay_port = port
            return 1

        KSR._mock_data["tm"]["t_relay_to_proto_addr"] = t_relay_to_proto_addr

        # indicate the call is not from a carrier ip address
        KSR._mock_data["permissions"]["allow_source_address"] = 0
        # indicate the call is coming from an asterisk ip
        KSR._mock_data["dispatcher"]["ds_is_from_list"] = 1
        # indicate that the ip found in the home parameter is not local
        KSR._mock_data[""]["is_myself"] = False

        k = kamailio.kamailio()
        k.kamailioDB = TestKamailioDatabase.TestKamailioDatabase()


        k.ksr_request_route(None)

        self.assertEqual(ksr_utils.pvar_get("$ru"), "sip:AgentSmit@agent.openline.ai",
                         "RURI not set to expected destination")  # add assertion here
        self.assertTrue(ksr_relay_called, "Call was not routed!")
        self.assertEqual(ksr_relay_destination, "127.0.0.1", "Call routed to wrong destination ip")
        self.assertEqual(ksr_relay_port, 5090, "Call routed to wrong destination port")


if __name__ == '__main__':
    unittest.main()
