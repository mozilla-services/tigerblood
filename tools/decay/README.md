# Decay lambda function

This AWS lambda function will increase the reputation of every single entry in the database by the `DECAY_RATE` set in the config file every time it is run.

Requires the env vars `DECAY_RATE` and `DB_DSN`.
