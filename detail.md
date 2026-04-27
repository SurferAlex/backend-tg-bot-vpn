Шаг 5 — Добавить vpnapi в /opt/vpnplatform/docker-compose.yml
Чтобы proxy_pass http://vpnapi:8080 заработал, в этом же compose должны появиться postgres + vpnapi (и позже bot). Для этого нужно, чтобы код проекта лежал в /opt/vpnplatform (папки backend/ и bot/).

Скажи, как ты будешь переносить код на VPS: git clone или scp/rsync — и я дам следующий короткий шаг 1-в-1 под твой вариант.