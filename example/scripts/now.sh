#! /bin/bash
cat <<EOF
# HELP answer_to_everything Answer To Everything -- by shell
# TYPE answer_to_everything gauge
current_time{scope="system",env="prod"} $(date +%s)
EOF
