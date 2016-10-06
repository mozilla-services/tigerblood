import json
import psycopg2


def decay(event, context):
    with open('config.json') as config_file:
        config = json.load(config_file)
    conn = psycopg2.connect(config['DB_DSN'])
    cur = conn.cursor()
    cur.execute(
        """
        UPDATE reputation
            SET reputation = reputation + %s
        WHERE reputation <= 100 - %s""",
        (config['DECAY_RATE'],) * 2)
    conn.commit()
    cur.close()
    conn.close()
