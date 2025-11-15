# How to Run Development Instance

When you are testing changes locally you unfortunately have to re-use the same discord
bot token as would actually be running and accepting your production requests.

In order to isolate your local changes and limit the scope of servers you're accepting
requests for we can leverage the `is_enabled` flag in the `servers` table. By setting this
flag to false for all other servers in the database besides the one you want to do your
testing in you can limit your local instance of the bot to only the single enabled server.

```sql
UPDATE servers SET is_enabled = 0 WHERE server_name != 'your server name';
```

This _does_ require the inverse action to be done on the "main server" to prevent the main
instance from trying to hijack your requests.
