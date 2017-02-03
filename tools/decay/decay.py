import os
import psycopg2


UPDATE_REPUTATION_SQL = """
UPDATE reputation
SET reputation = reputation + %s
WHERE reputation <= 100 - %s
"""


def decay():
    with psycopg2.connect(os.environ['DB_DSN']) as conn:
        with conn.cursor() as cur:
            cur.execute(UPDATE_REPUTATION_SQL,
                        (os.environ['DECAY_RATE'],) * 2)
    conn.close()


def handler(event, context):
    decay()


if __name__ == '__main__':
    decay()
