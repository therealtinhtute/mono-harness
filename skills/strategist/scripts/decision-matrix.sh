#!/bin/bash
# Generate decision comparison matrix
# Usage: ./decision-matrix.sh "Option 1" "Option 2" "Option 3"

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SKILL_DIR="$(dirname "$SCRIPT_DIR")"

function show_usage() {
  cat << 'USAGE'
Usage: $(basename "$0") <option1> <option2> [option3]

Generate decision comparison matrix with weighted scoring

Arguments:
  option1, option2, option3    Options to compare (2-3 required)

Options:
  -h, --help       Show this help message
  -o, --output     Output file (default: .kit/reports/brainstorm/YYYYMMDD-decision.md)

Examples:
  ./decision-matrix.sh "REST API" "GraphQL"
  ./decision-matrix.sh "Monolith" "Microservices" "Serverless"

from therealTINHTUTE with love
USAGE
}

function prompt_criteria() {
  echo "Enter evaluation criteria (one per line, empty line to finish):"
  echo "Examples: Complexity, Cost, Time, Scalability, Maintainability"
  echo ""
  
  local criteria=()
  while true; do
    read -p "Criterion: " criterion
    if [ -z "$criterion" ]; then
      break
    fi
    criteria+=("$criterion")
  done
  
  printf '%s\n' "${criteria[@]}"
}

function prompt_weights() {
  local criteria=("$@")
  
  echo ""
  echo "Enter weights for each criterion (1-10):"
  
  local weights=()
  for criterion in "${criteria[@]}"; do
    read -p "$criterion weight (1-10): " weight
    weights+=("$weight")
  done
  
  printf '%s\n' "${weights[@]}"
}

function prompt_scores() {
  local option="$1"
  shift
  local criteria=("$@")
  
  echo ""
  echo "Score '$option' for each criterion (1-10):"
  
  local scores=()
  for criterion in "${criteria[@]}"; do
    read -p "$criterion score (1-10): " score
    scores+=("$score")
  done
  
  printf '%s\n' "${scores[@]}"
}

function calculate_weighted_score() {
  local scores=("${!1}")
  local weights=("${!2}")
  
  local total=0
  local weight_sum=0
  
  for i in "${!scores[@]}"; do
    local weighted=$((scores[i] * weights[i]))
    total=$((total + weighted))
    weight_sum=$((weight_sum + weights[i]))
  done
  
  echo $((total * 100 / weight_sum))
}

function generate_matrix() {
  local options=("$@")
  
  echo "## Decision Matrix"
  echo ""
  
  # Get criteria
  local criteria=($(prompt_criteria))
  
  if [ ${#criteria[@]} -eq 0 ]; then
    echo "❌ No criteria provided"
    return 1
  fi
  
  # Get weights
  local weights=($(prompt_weights "${criteria[@]}"))
  
  # Get scores for each option
  declare -A all_scores
  for option in "${options[@]}"; do
    local scores=($(prompt_scores "$option" "${criteria[@]}"))
    all_scores["$option"]="${scores[*]}"
  done
  
  # Generate table header
  echo "| Criterion | Weight | ${options[*]} |"
  echo "|-----------|--------|$(printf '%s|' "${options[@]/#/--------}")"|
  
  # Generate table rows
  for i in "${!criteria[@]}"; do
    echo -n "| ${criteria[$i]} | ${weights[$i]} |"
    for option in "${options[@]}"; do
      local scores=(${all_scores["$option"]})
      echo -n " ${scores[$i]} |"
    done
    echo ""
  done
  
  # Calculate weighted scores
  echo ""
  echo "### Weighted Scores"
  echo ""
  
  declare -A final_scores
  for option in "${options[@]}"; do
    local scores=(${all_scores["$option"]})
    local score=$(calculate_weighted_score scores[@] weights[@])
    final_scores["$option"]=$score
    echo "- **$option**: $score/100"
  done
  
  # Recommendation
  echo ""
  echo "### Recommendation"
  echo ""
  
  local best_option=""
  local best_score=0
  
  for option in "${options[@]}"; do
    if [ ${final_scores["$option"]} -gt $best_score ]; then
      best_score=${final_scores["$option"]}
      best_option="$option"
    fi
  done
  
  echo "✅ **Recommended**: $best_option (score: $best_score/100)"
  echo ""
}

function main() {
  local options=()
  local output_file=""

  # Parse arguments
  while [[ $# -gt 0 ]]; do
    case $1 in
      -h|--help)
        show_usage
        exit 0
        ;;
      -o|--output)
        output_file="$2"
        shift 2
        ;;
      *)
        options+=("$1")
        shift
        ;;
    esac
  done

  if [ ${#options[@]} -lt 2 ]; then
    echo "❌ Error: at least 2 options required"
    show_usage
    exit 1
  fi

  # Set default output file
  if [ -z "$output_file" ]; then
    local date=$(date +%Y%m%d)
    output_file=".kit/reports/brainstorm/${date}-decision.md"
    mkdir -p "$(dirname "$output_file")"
  fi

  echo "🎯 Decision Matrix Generator"
  echo "============================"
  echo ""
  echo "Options: ${options[*]}"
  echo ""

  # Generate matrix
  local matrix=$(generate_matrix "${options[@]}")

  # Save report
  {
    echo "---"
    echo "title: Decision Matrix"
    echo "description: Comparison of ${options[*]}"
    echo "status: completed"
    echo "created: $(date +%Y-%m-%d)"
    echo "tags: [decision, strategy]"
    echo "---"
    echo ""
    echo "# Decision Matrix"
    echo ""
    echo "**Options**: ${options[*]}"
    echo "**Generated**: $(date)"
    echo ""
    echo "$matrix"
    echo ""
    echo "from therealTINHTUTE with love"
  } > "$output_file"

  echo ""
  echo "✅ Decision matrix saved: $output_file"
}

main "$@"
