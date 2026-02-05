#!/bin/bash

# 日志归档脚本
# 自动归档和清理Elasticsearch旧日志

set -e

# 配置
ELASTICSEARCH_HOST="${ELASTICSEARCH_HOST:-localhost:9200}"
ARCHIVE_DAYS=${ARCHIVE_DAYS:-30}
DELETE_DAYS=${DELETE_DAYS:-90}
S3_BUCKET="${S3_BUCKET:-enterprise-platform-logs}"
S3_REGION="${S3_REGION:-us-west-2}"

# 颜色
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查Elasticsearch连接
check_elasticsearch() {
    log_info "Checking Elasticsearch connection..."
    if ! curl -sf "http://${ELASTICSEARCH_HOST}/_cluster/health" > /dev/null; then
        log_error "Cannot connect to Elasticsearch at ${ELASTICSEARCH_HOST}"
        exit 1
    fi
    log_info "Elasticsearch is reachable"
}

# 获取需要归档的索引
get_indices_to_archive() {
    local cutoff_date=$(date -d "${ARCHIVE_DAYS} days ago" +%Y.%m.%d)
    log_info "Finding indices older than ${cutoff_date}"

    curl -s "http://${ELASTICSEARCH_HOST}/_cat/indices/app-logs-*?h=index" | \
        awk -v cutoff="$cutoff_date" '
            {
                if (match($0, /[0-9]{4}\.[0-9]{2}\.[0-9]{2}/)) {
                    date = substr($0, RSTART, RLENGTH)
                    gsub(/\./, "", date)
                    gsub(/\./, "", cutoff)
                    if (date < cutoff) print $0
                }
            }
        '
}

# 归档索引到S3
archive_index() {
    local index=$1
    local snapshot_name="snapshot-${index}-$(date +%Y%m%d%H%M%S)"
    local snapshot_repo="s3_backup"

    log_info "Archiving index: ${index}"

    # 注册S3快照仓库（如果未注册）
    curl -X PUT "http://${ELASTICSEARCH_HOST}/_snapshot/${snapshot_repo}" \
        -H 'Content-Type: application/json' \
        -d "{
            \"type\": \"s3\",
            \"settings\": {
                \"bucket\": \"${S3_BUCKET}\",
                \"region\": \"${S3_REGION}\",
                \"base_path\": \"elasticsearch-snapshots\"
            }
        }" 2>/dev/null

    # 创建快照
    curl -X PUT "http://${ELASTICSEARCH_HOST}/_snapshot/${snapshot_repo}/${snapshot_name}?wait_for_completion=true" \
        -H 'Content-Type: application/json' \
        -d "{
            \"indices\": \"${index}\",
            \"ignore_unavailable\": true,
            \"include_global_state\": false
        }" 2>/dev/null

    if [ $? -eq 0 ]; then
        log_info "Successfully archived ${index} to S3"
    else
        log_error "Failed to archive ${index}"
        return 1
    fi
}

# 删除旧索引
delete_old_indices() {
    local cutoff_date=$(date -d "${DELETE_DAYS} days ago" +%Y.%m.%d)
    log_info "Deleting indices older than ${cutoff_date}"

    local indices=$(curl -s "http://${ELASTICSEARCH_HOST}/_cat/indices/app-logs-*?h=index" | \
        awk -v cutoff="$cutoff_date" '
            {
                if (match($0, /[0-9]{4}\.[0-9]{2}\.[0-9]{2}/)) {
                    date = substr($0, RSTART, RLENGTH)
                    gsub(/\./, "", date)
                    gsub(/\./, "", cutoff)
                    if (date < cutoff) print $0
                }
            }
        ')

    for index in $indices; do
        log_info "Deleting index: ${index}"
        curl -X DELETE "http://${ELASTICSEARCH_HOST}/${index}" 2>/dev/null
        if [ $? -eq 0 ]; then
            log_info "Successfully deleted ${index}"
        else
            log_error "Failed to delete ${index}"
        fi
    done
}

# 强制合并段
force_merge_old_indices() {
    local cutoff_date=$(date -d "7 days ago" +%Y.%m.%d)
    log_info "Force merging indices older than ${cutoff_date}"

    local indices=$(curl -s "http://${ELASTICSEARCH_HOST}/_cat/indices/app-logs-*?h=index" | \
        awk -v cutoff="$cutoff_date" '
            {
                if (match($0, /[0-9]{4}\.[0-9]{2}\.[0-9]{2}/)) {
                    date = substr($0, RSTART, RLENGTH)
                    gsub(/\./, "", date)
                    gsub(/\./, "", cutoff)
                    if (date < cutoff) print $0
                }
            }
        ')

    for index in $indices; do
        log_info "Force merging index: ${index}"
        curl -X POST "http://${ELASTICSEARCH_HOST}/${index}/_forcemerge?max_num_segments=1" 2>/dev/null &
    done

    wait
    log_info "Force merge completed"
}

# 生成归档报告
generate_report() {
    log_info "Generating archive report..."

    local total_indices=$(curl -s "http://${ELASTICSEARCH_HOST}/_cat/indices/app-logs-*?h=index" | wc -l)
    local total_size=$(curl -s "http://${ELASTICSEARCH_HOST}/_cat/indices/app-logs-*?h=store.size" | \
        awk '{sum += $1} END {print sum/1024/1024/1024 " GB"}')

    cat <<EOF
========================================
Elasticsearch Log Archive Report
Generated: $(date)
========================================

Total Indices: ${total_indices}
Total Size: ${total_size}

Archive Policy:
- Archive after: ${ARCHIVE_DAYS} days
- Delete after: ${DELETE_DAYS} days
- S3 Bucket: ${S3_BUCKET}

========================================
EOF
}

# 主函数
main() {
    log_info "Starting log archive job..."

    check_elasticsearch

    # 强制合并旧索引（提高压缩率）
    force_merge_old_indices

    # 归档索引到S3
    if [ -n "${S3_BUCKET}" ]; then
        local indices=$(get_indices_to_archive)
        for index in $indices; do
            archive_index "$index" || true
        done
    else
        log_warn "S3_BUCKET not set, skipping archive to S3"
    fi

    # 删除旧索引
    delete_old_indices

    # 生成报告
    generate_report

    log_info "Log archive job completed"
}

# 运行主函数
main "$@"
