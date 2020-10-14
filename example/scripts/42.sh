#! /bin/bash
cat <<EOF
# HELP answer_to_everything Answer To Everything -- by shell
# TYPE answer_to_everything gauge
answer_to_everything{scope="universe",env="prod"} 42
EOF