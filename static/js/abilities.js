const Abilities = {
    chart: null,

    async load() {
        try {
            const abilities = await App.api('/api/abilities');
            this.renderList(abilities);
        } catch (err) {
            App.toast('加载能力列表失败: ' + err.message, 'error');
        }
    },

    renderList(abilities) {
        const container = document.getElementById('abilities-list');
        const detail = document.getElementById('ability-detail');
        detail.hidden = true;
        container.hidden = false;

        if (!abilities || abilities.length === 0) {
            container.innerHTML = '<div class="empty-state"><div class="empty-state-icon">🎯</div><p>还没有能力，点击"+ 新增能力"创建第一个</p></div>';
            return;
        }

        container.innerHTML = abilities.map(a => {
            const decayInfo = a.days_since_last_activity > 0
                ? `<span style="color:var(--warning)">衰减中 (${a.days_since_last_activity.toFixed(1)}天)</span>`
                : '';
            return `
                <div class="ability-row">
                    <div class="ability-row-info">
                        <h4>${this.escapeHtml(a.name)}</h4>
                        <p>
                            当前值: ${App.formatNumber(a.current_value)}
                            ${a.effective_value !== a.current_value ? ` / 有效值: ${App.formatNumber(a.effective_value)}` : ''}
                            | 基础: ${App.formatNumber(a.base_value)} ${decayInfo}
                        </p>
                    </div>
                    <div class="row-actions">
                        <button class="btn btn-secondary btn-sm" onclick="Abilities.showDetail(${a.id})">详情</button>
                        <button class="btn btn-secondary btn-sm" onclick="Abilities.showForm(${a.id})">编辑</button>
                        <button class="btn btn-danger btn-sm" onclick="Abilities.deleteAbility(${a.id})">删除</button>
                    </div>
                </div>
            `;
        }).join('');
    },

    backToList() {
        document.getElementById('abilities-list').hidden = false;
        document.getElementById('ability-detail').hidden = true;
        if (this.chart) {
            this.chart.destroy();
            this.chart = null;
        }
        this.load();
    },

    async showDetail(id) {
        document.getElementById('abilities-list').hidden = true;
        const detail = document.getElementById('ability-detail');
        detail.hidden = false;

        try {
            const result = await App.api(`/api/abilities/${id}/history`);
            const a = result.ability;
            const points = result.points || [];

            document.getElementById('ability-detail-content').innerHTML = `
                <div class="detail-header">
                    <h3>${this.escapeHtml(a.name)}</h3>
                    <p>${this.escapeHtml(a.description) || '暂无描述'}</p>
                </div>
                <div class="detail-stats">
                    <div class="detail-stat">
                        <div class="detail-stat-label">当前值</div>
                        <div class="detail-stat-value">${App.formatNumber(a.current_value)}</div>
                    </div>
                    <div class="detail-stat">
                        <div class="detail-stat-label">有效值（衰减后）</div>
                        <div class="detail-stat-value">${App.formatNumber(a.effective_value)}</div>
                    </div>
                    <div class="detail-stat">
                        <div class="detail-stat-label">基础值</div>
                        <div class="detail-stat-value">${App.formatNumber(a.base_value)}</div>
                    </div>
                    <div class="detail-stat">
                        <div class="detail-stat-label">成长倍率</div>
                        <div class="detail-stat-value">${App.formatNumber(a.growth_rate)}x</div>
                    </div>
                    <div class="detail-stat">
                        <div class="detail-stat-label">日衰减率</div>
                        <div class="detail-stat-value">${App.formatNumber(a.decay_rate)}%</div>
                    </div>
                    <div class="detail-stat">
                        <div class="detail-stat-label">上次活动</div>
                        <div class="detail-stat-value" style="font-size:14px">${a.last_activity_at ? App.timeAgo(a.last_activity_at) : '从未'}</div>
                    </div>
                </div>
                <div class="chart-container">
                    <canvas id="ability-chart"></canvas>
                </div>
                <div class="detail-actions">
                    <button class="btn btn-primary" onclick="Abilities.showForm(${a.id})">编辑能力</button>
                    <button class="btn btn-danger" onclick="Abilities.deleteAbility(${a.id})">删除能力</button>
                </div>
            `;

            // Render chart
            this.renderChart(points, a.name);

        } catch (err) {
            App.toast('加载能力详情失败: ' + err.message, 'error');
        }
    },

    renderChart(points, abilityName) {
        const canvas = document.getElementById('ability-chart');
        if (!canvas) return;

        if (this.chart) this.chart.destroy();

        const labels = points.map(p => {
            const d = new Date(p.date + (p.date.includes('T') ? '' : 'Z'));
            return isNaN(d.getTime()) ? p.date : d.toLocaleDateString('zh-CN');
        });
        const values = points.map(p => p.value);

        this.chart = new Chart(canvas, {
            type: 'line',
            data: {
                labels: labels,
                datasets: [{
                    label: abilityName,
                    data: values,
                    borderColor: '#4f46e5',
                    backgroundColor: 'rgba(79,70,229,0.1)',
                    tension: 0.2,
                    fill: true,
                    pointRadius: 5,
                    pointHoverRadius: 7,
                    pointBackgroundColor: '#4f46e5'
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                plugins: {
                    legend: { display: false },
                    tooltip: {
                        callbacks: {
                            label: (ctx) => `${abilityName}: ${App.formatNumber(ctx.parsed.y)}`
                        }
                    }
                },
                scales: {
                    y: { beginAtZero: false, title: { display: true, text: '能力值' } },
                    x: { title: { display: true, text: '日期' } }
                }
            }
        });
    },

    async showForm(id) {
        let ability = { name: '', description: '', base_value: 0, current_value: 0, growth_rate: 1.0, decay_rate: 0.5 };
        let isEdit = false;

        if (id) {
            isEdit = true;
            try {
                ability = await App.api(`/api/abilities/${id}`);
            } catch (err) {
                App.toast('加载能力失败', 'error');
                return;
            }
        }

        Modal.show(isEdit ? '编辑能力' : '新建能力', `
            <div class="form-group">
                <label>名称 *</label>
                <input type="text" id="form-name" value="${this.escapeHtml(ability.name)}" class="input">
            </div>
            <div class="form-group">
                <label>描述</label>
                <textarea id="form-desc" class="input">${this.escapeHtml(ability.description || '')}</textarea>
            </div>
            <div class="form-row">
                <div class="form-group">
                    <label>基础值 *</label>
                    <input type="number" id="form-base" value="${ability.base_value}" step="1" class="input">
                </div>
                <div class="form-group">
                    <label>当前值</label>
                    <input type="number" id="form-current" value="${ability.current_value || ability.base_value}" step="1" class="input">
                </div>
            </div>
            <div class="form-row">
                <div class="form-group">
                    <label>成长倍率 (默认 1.0)</label>
                    <input type="number" id="form-growth" value="${ability.growth_rate}" step="0.1" class="input">
                </div>
                <div class="form-group">
                    <label>日衰减率 % (默认 0.5)</label>
                    <input type="number" id="form-decay" value="${ability.decay_rate}" step="0.1" class="input">
                </div>
            </div>
            <div class="form-actions">
                <button class="btn btn-secondary" onclick="Modal.hide()">取消</button>
                <button class="btn btn-primary" id="form-save">保存</button>
            </div>
        `);

        document.getElementById('form-save').onclick = async () => {
            const data = {
                name: document.getElementById('form-name').value,
                description: document.getElementById('form-desc').value,
                base_value: parseFloat(document.getElementById('form-base').value) || 0,
                current_value: parseFloat(document.getElementById('form-current').value) || 0,
                growth_rate: parseFloat(document.getElementById('form-growth').value) || 1.0,
                decay_rate: parseFloat(document.getElementById('form-decay').value) || 0.5
            };
            if (!data.name) { App.toast('请输入名称', 'warning'); return; }

            try {
                if (isEdit) {
                    await App.api(`/api/abilities/${id}`, { method: 'PUT', body: JSON.stringify(data) });
                } else {
                    await App.api('/api/abilities', { method: 'POST', body: JSON.stringify(data) });
                }
                Modal.hide();
                App.toast(isEdit ? '能力已更新' : '能力已创建');
                this.load();
            } catch (err) {
                App.toast('保存失败: ' + err.message, 'error');
            }
        };
    },

    async deleteAbility(id) {
        if (!confirm('确定要删除这个能力吗？相关效果和快照也会被删除。')) return;
        try {
            await App.api(`/api/abilities/${id}`, { method: 'DELETE' });
            App.toast('能力已删除');
            this.load();
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
