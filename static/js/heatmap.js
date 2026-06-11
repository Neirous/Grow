const Heatmap = {
    render(containerId, data) {
        const container = document.getElementById(containerId);
        if (!container) return;

        const { heatmap, currentStreak, longestStreak } = data;
        if (!heatmap || heatmap.length === 0) {
            container.innerHTML = '';
            return;
        }

        // Streak info bar
        let html = '<div class="streak-bar">';
        html += `<span class="streak-badge">🔥 当前连续: <strong>${currentStreak}</strong> 天</span>`;
        html += `<span class="streak-badge">🏆 历史最长: <strong>${longestStreak}</strong> 天</span>`;
        html += '</div>';

        // Heatmap grid (GitHub-style)
        html += '<div class="heatmap-grid">';
        html += '<div class="heatmap-months">';

        // Month labels
        const months = ['1月','2月','3月','4月','5月','6月','7月','8月','9月','10月','11月','12月'];
        let lastMonth = -1;
        let monthLabels = '';
        heatmap.forEach((day, idx) => {
            const d = new Date(day.date + 'T00:00:00');
            const month = d.getMonth();
            if (month !== lastMonth) {
                const pos = Math.floor(idx / 7) * 14;
                monthLabels += `<span style="grid-column:${Math.max(1, Math.floor(idx / 7) + 1)}">${months[month]}</span>`;
                lastMonth = month;
            }
        });
        html += `<div class="heatmap-month-labels">${monthLabels}</div>`;

        // Day cells
        html += '<div class="heatmap-cells">';
        heatmap.forEach(day => {
            let level = 0;
            if (day.count >= 5) level = 4;
            else if (day.count >= 3) level = 3;
            else if (day.count >= 2) level = 2;
            else if (day.count >= 1) level = 1;

            const d = new Date(day.date + 'T00:00:00');
            const tooltip = `${d.toLocaleDateString('zh-CN')}: ${day.count} 次活动`;
            html += `<div class="heatmap-cell level-${level}" title="${tooltip}"></div>`;
        });
        html += '</div><div class="heatmap-legend">';
        html += '<span>少</span>';
        for (let i = 0; i <= 4; i++) {
            html += `<div class="heatmap-cell level-${i}" style="width:12px;height:12px"></div>`;
        }
        html += '<span>多</span></div></div></div>';

        container.innerHTML = html;
    }
};
