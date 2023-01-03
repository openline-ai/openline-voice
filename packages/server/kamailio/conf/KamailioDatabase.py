import psycopg2

class KamailioDatabase:
    connection = None
    conn_string = None

    def __init__(self, host: str, database: str, user: str, password:str):
        self.conn_string = "host='%s' dbname='%s' user='%s' password='%s'" % (host, database, user, password)
        self.connection = psycopg2.connect(self.conn_string)
        self.connection.set_session(autocommit=True)

    def test_connection(self):
        try:
            cur = self.connection.cursor()
            cur.execute('SELECT 1')
        except psycopg2.OperationalError:
            self.connection = psycopg2.connect(self.conn_string)

    def lookup_carrier(self, carrier:str):
        self.test_connection()
        with self.connection.cursor() as cur:

            cur.execute("SELECT username, ha1, domain FROM openline_carrier WHERE carrier_name=%s",
                        (carrier,))
            record = cur.fetchone()
            if record is not None:
                return {"username": record[0],
                        "ha1": record[1],
                        "domain": record[2]}
        return None

    def find_sipuri_mapping(self, sipuri:str):
        self.test_connection()
        with self.connection.cursor() as cur:

            cur.execute("SELECT e164, carrier_name FROM openline_number_mapping WHERE sipuri=%s",
                        (sipuri,))
            record = cur.fetchone()
            if record is not None:
                return {"e164": record[0],
                        "carrier": record[1]
                        }
        return None

    def find_e164_mapping(self, e164:str, carrier:str):
        self.test_connection()
        with self.connection.cursor() as cur:
            cur.execute("SELECT sipuri FROM openline_number_mapping WHERE e164=%s AND carrier_name=%s", (e164, carrier))
            record = cur.fetchone()
            if record is not None:
                return {"sipuri": record[0]}
        return None

