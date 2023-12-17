docker run -d --cgroupns host \
              --pid host \
              --name datadog \
              -v /var/run/docker.sock:/var/run/docker.sock:ro \
              -v /proc/:/host/proc/:ro \
              -v /sys/fs/cgroup/:/host/sys/fs/cgroup:ro \
              -p 127.0.0.1:8126:8126/tcp \
              -e DD_API_KEY=<<>> \
              -e DD_APM_ENABLED=true \
              -e DD_SITE=us5.datadoghq.com \
              --network=common \
              gcr.io/datadoghq/agent:latest
