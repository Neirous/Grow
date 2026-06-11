const Logs = {
    currentOffset: 0,
    currentFilters: {},

    async load(filters = {}) {
        this.currentFilters = filters;
        this.currentOffset = 0;

        try {
            // Load filter dropdowns
            const [activities, abilities] = await Promise.all([
                App.api('/api/activities'),
                App.api('/api/abilities')
            ]);

            App.fetchSelectOptions('log-filter-activity', activities, 'id', 'name', '全部活动');
            App.fetchSelectOptions('log-filter-ability', abilities, 'id', 'name', '全部能力');

            // Apply saved filter values
            if (filters.activity_id) document.getElementById('log-filter-activity').value = filters.activity_id;
            if (filters.ability_id) document.getElementById('log-filter-ability').value = filters.ability_id;
        } catch (err) {
            App.toast('加载筛选项失败', 'error');
        }

        document.getElementById('log-filter-apply').onclick = () => {
            const filters = {};
            const actId = document.getElementById('log-filter-activity').value;
            const abiId = document.getElementById('log-filter-ability').value;
            if (actId) filters.activity_id = actId;
            if (abiId) filters.ability_id = abiId;
            this.load(filters);
        };

        await this.fetchLogs();
    },

    async fetchLogs() {
        const params = new URLSearchParams();
        params.set('limit', '20');
        params.set('offset', String(this.currentOffset));
        if (this.currentFilters.activity_id) params.set('activity_id', this.currentFilters.activity_id);
        if (this.currentFilters.ability_id) params.set('ability_id', this.currentFilters.ability_id);

        const container = document.getElementById('logs-list');
        const loadMoreBtn = document.getElementById('logs-load-more');

        try {
            const logs = await App.api(`/api/logs?${params.toString()}`);

            if (this.currentOffset === 0) {
                container.innerHTML = '';
            }

            if (!logs || logs.length === 0) {
                if (this.currentOffset === 0) {
                    container.innerHTML = '<div class="empty-state"><div class="empty-state-icon">📋</div><p>暂无活动日志</p></div>';
                }
                loadMoreBtn.style.display = 'none';
                return;
            }

            container.innerHTML += logs.map(log => `
                <div class="log-detailed">
                    <div class="log-detailed-header">
                        <span class="log-detailed-name">${this.escapeHtml(log.activity_name)}</span>
                        <span class="log-detailed-time">${App.formatDate(log.completed_at)}</span>
                    </div>
                    <div class="log-detailed-changes">
                        ${(log.snapshots || []).map(s => {
                            const diff = s.new_value - s.old_value;
                            const pct = (diff / s.old_value * 100).toFixed(2);
                            return `
                                <span class="log-change log-change-up">
                                    ${this.escapeHtml(s.ability_name)}:
                                    ${App.formatNumber(s.old_value)} → ${App.formatNumber(s.new_value)}
                                    (+${pct}%)
                                </span>
                            `;
                        }).join('')}
                    </div>
                    ${log.note ? `<div style="margin-top:6px;font-size:12px;color:var(--text-secondary)">📝 ${this.escapeHtml(log.note)}</div>` : ''}
                    <button class="btn btn-danger btn-xs" onclick="Logs.deleteLog(${log.id})" style="margin-top:8px">删除此记录</button>
                </div>
            `).join('');

            loadMoreBtn.style.display = logs.length < 20 ? 'none' : 'block';
            loadMoreBtn.onclick = () => {
                this.currentOffset += 20;
                this.fetchLogs();
            };

        } catch (err) {
            App.toast('加载日志失败: ' + err.message, 'error');
        }
    },

    async deleteLog(id) {
        if (!confirm('确定要删除这条日志记录吗？')) return;
        try {
            await App.api(`/api/logs/${id}`, { method: 'DELETE' });
            App.toast('日志已删除');
            this.currentOffset = 0;
            this.fetchLogs();
        } catch (err) {
            App.toast('删除失败: ' + err.message, 'error');
        }
    },

    escapeHtml(str) {
        const div = document.createElement('div');
        div.textContent = str || '';
        return div.innerHTML;
    }
};
