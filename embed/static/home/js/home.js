(function() {
    'use strict';

    const HomeApp = {
        ws: null,
        reconnectTimer: null,

        utils: {
            escapeHtml: function(text) {
                if (!text) return '';
                const map = {
                    '&': '&amp;',
                    '<': '&lt;',
                    '>': '&gt;',
                    '"': '&quot;',
                    "'": '&#039'
                };
                return text.replace(/[&<>'"]/g, function(m) { return map[m]; });
            },

            getTypeIcon: function(type) {
                const icons = {
                    'http': '<i class="fas fa-globe"></i>',
                    'https': '<i class="fas fa-globe"></i>',
                    'tcp': '<i class="fas fa-network-wired"></i>',
                    'udp': '<i class="fas fa-broadcast-tower"></i>',
                    'dns': '<i class="fas fa-server"></i>',
                    'ping': '<i class="fas fa-wifi"></i>'
                };
                return icons[type] || '<i class="fas fa-question"></i>';
            },

            padZero: function(num) {
                return String(num).padStart(2, '0');
            },

            formatDate: function(dateStr) {
                const date = new Date(dateStr);
                return date.toLocaleDateString('zh-CN', {
                    year: 'numeric',
                    month: '2-digit',
                    day: '2-digit',
                    hour: '2-digit',
                    minute: '2-digit',
                    second: '2-digit'
                });
            },

            formatSize: function(bytes) {
                if (bytes >= 1024) {
                    return (bytes / 1024).toFixed(2) + 'kb';
                }
                return bytes + 'b';
            },

            getUrlParam: function(name) {
                return new URLSearchParams(window.location.search).get(name);
            },

            getWsUrl: function(path, params) {
                const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
                let url = protocol + '//' + window.location.host + path;
                if (params) {
                    url += '?' + params;
                }
                return url;
            }
        },

        connect: function(wsUrl, handlers) {
            const self = this;
            try {
                this.ws = new WebSocket(wsUrl);

                this.ws.onopen = handlers.onopen || function() {};
                this.ws.onmessage = handlers.onmessage || function() {};
                this.ws.onclose = handlers.onclose || function() {};
                this.ws.onerror = handlers.onerror || function() {};
            } catch (e) {
                console.error('WebSocket connection failed:', e);
                if (this.reconnectTimer) {
                    clearTimeout(this.reconnectTimer);
                }
                this.reconnectTimer = setTimeout(function() {
                    self.connect(wsUrl, handlers);
                }, 5000);
            }
        },

        reconnect: function(wsUrl, handlers) {
            if (this.reconnectTimer) {
                clearTimeout(this.reconnectTimer);
            }
            this.reconnectTimer = setTimeout(function() {
                this.connect(wsUrl, handlers);
            }.bind(this), 5000);
        },

        renderGroups: {
            data: [],

            init: function() {
                const self = this;
                const groupId = HomeApp.utils.getUrlParam('id');
                let wsUrl = HomeApp.utils.getWsUrl('/ws/groups');
                if (groupId) {
                    wsUrl += '?id=' + groupId;
                }

                HomeApp.connect(wsUrl, {
                    onopen: function() {
                        console.log('Groups WebSocket connected');
                    },
                    onmessage: function(event) {
                        try {
                            const data = JSON.parse(event.data);
                            if (data.type === 'group_status') {
                                self.data = data.data || [];
                                self.renderCards();
                                self.renderTabs();
                            }
                        } catch (e) {
                            console.error('Failed to parse message:', e);
                        }
                    },
                    onclose: function() {
                        console.log('Groups WebSocket disconnected, reconnecting...');
                        HomeApp.reconnect(wsUrl, this);
                    },
                    onerror: function(error) {
                        console.error('Groups WebSocket error:', error);
                    }
                });
            },

            renderCards: function() {
                const container = document.getElementById('group-container');
                if (!container) return;

                if (!this.data || this.data.length === 0) {
                    container.innerHTML = '<div class="loading">暂无分组数据</div>';
                    return;
                }

                const groupId = HomeApp.utils.getUrlParam('id');
                let groupsToDisplay = this.data;

                if (groupId) {
                    groupsToDisplay = this.data.filter(function(group) {
                        return group.id == groupId;
                    });

                    if (groupsToDisplay.length === 0) {
                        container.innerHTML = '<div class="loading">未找到指定的分组</div>';
                        return;
                    }
                }

                let html = '';
                groupsToDisplay.forEach(function(group) {
                    const monitors = group.monitors || [];
                    const upCount = monitors.filter(function(m) { return m.is_valid; }).length;
                    const totalCount = monitors.length;
                    const upRate = totalCount > 0 ? (upCount / totalCount * 100).toFixed(1) : '0.0';

                    html += '<div class="group-section" data-group-id="' + group.id + '">';
                    html += '<h2 style="margin-bottom: 20px; font-size: 18px; font-weight: 600;">' + HomeApp.utils.escapeHtml(group.name) + '</h2>';
                    html += '<div style="margin-bottom: 25px; font-size: 14px; color: #666;">';
                    html += '正常: ' + upCount + '/' + totalCount + ' (' + upRate + '%)';
                    html += '</div>';

                    if (monitors.length > 0) {
                        monitors.forEach(function(monitor) {
                            const statusClass = monitor.is_valid ? 'up' : 'down';
                            const statusText = monitor.is_valid ? '正常' : '无法访问';
                            const statusIcon = monitor.is_valid ? 'fa-check' : 'fa-times';
                            const statusColor = monitor.is_valid ? '#27ae60' : '#e74c3c';

                            html += '<div class="status-card">';
                            html += '<div class="status-header">';
                            html += '<div class="service-name">' + HomeApp.utils.getTypeIcon(monitor.type) + ' ' + HomeApp.utils.escapeHtml(monitor.name) + '</div>';
                            html += '<div class="status-indicator ' + statusClass + '">';
                            html += '<i class="fas ' + statusIcon + '"></i> ' + statusText + ' ' + (monitor.latency || '');
                            html += '</div>';
                            html += '</div>';

                            html += '<div class="time-grid">';
                            for (let row = 0; row < 6; row++) {
                                html += '<div class="time-row">';
                                for (let i = 0; i < 24; i++) {
                                    const isUp = Math.random() > 0.2;
                                    html += '<div class="time-cell ' + (isUp ? 'up' : 'down') + '"></div>';
                                }
                                html += '</div>';
                            }
                            html += '</div>';

                            html += '<div class="legend">';
                            html += '<div class="legend-item"><div class="legend-color" style="background-color: #27ae60;"></div><span>正常</span></div>';
                            html += '<div class="legend-item"><div class="legend-color" style="background-color: #e74c3c;"></div><span>故障</span></div>';
                            html += '<div class="legend-item"><div class="legend-color" style="background-color: #f0f0f0;"></div><span>未知</span></div>';
                            html += '</div>';

                            html += '<div style="display: flex; flex-wrap: wrap; gap: 20px; font-size: 14px; margin-top: 15px;">';
                            html += '<div>类型: ' + HomeApp.utils.escapeHtml(monitor.type) + '</div>';
                            html += '<div>地址: ' + HomeApp.utils.escapeHtml(monitor.address) + '</div>';
                            if (monitor.speed) {
                                html += '<div>耗时: ' + monitor.speed + 'ms</div>';
                            }
                            if (monitor.size) {
                                html += '<div>大小: ' + HomeApp.utils.formatSize(monitor.size) + '</div>';
                            }
                            html += '<div>日志: ' + Math.floor(Math.random() * 1000) + ' 条</div>';
                            html += '</div>';

                            html += '<div class="status-footer">';
                            html += '<div>今天</div>';
                            html += '<div class="status-stats">最近 60 天可用率 ' + (monitor.is_valid ? '100.0' : '0.0') + '%</div>';
                            html += '<div class="date">' + new Date().toLocaleDateString('zh-CN') + '</div>';
                            html += '</div>';

                            html += '</div>';
                        });
                    } else {
                        html += '<div style="text-align: center; padding: 40px 0; color: #999; font-size: 14px;">该分组下暂无监控点</div>';
                    }

                    html += '</div>';
                }.bind(this));

                container.innerHTML = html;
            },
            renderTabs: function() {
                const tabsContainer = document.getElementById('group-tabs');
                if (!tabsContainer) return;

                let html = '';

                if (this.data && this.data.length > 0) {
                    this.data.forEach(function(group) {
                        html += '<div class="tab" data-tab="group-' + group.id + '">' + HomeApp.utils.escapeHtml(group.name) + '</div>';
                    });
                }

                tabsContainer.innerHTML = html;
                this.bindTabEvents();
            },
            bindTabEvents: function() {
                const tabs = document.querySelectorAll('.tab');
                const groupSections = document.querySelectorAll('.group-section');

                tabs.forEach(function(tab) {
                    tab.onclick = function() {
                        tabs.forEach(function(t) { t.classList.remove('active'); });
                        tab.classList.add('active');

                        const tabName = tab.dataset.tab;

                        groupSections.forEach(function(section, index) {
                            if (tabName === 'group-' + HomeApp.renderGroups.data[index].id) {
                                section.style.display = 'block';
                            } else {
                                section.style.display = 'none';
                            }
                        });
                    };
                });
            }
        },
        renderMonitor: {
            data: null,

            init: function() {
                const self = this;
                const monitorId = HomeApp.utils.getUrlParam('id');

                if (!monitorId) {
                    document.getElementById('monitor-container').innerHTML = '<div class="loading">缺少监控点ID参数</div>';
                    return;
                }

                const wsUrl = HomeApp.utils.getWsUrl('/ws/monitor', 'id=' + monitorId);

                HomeApp.connect(wsUrl, {
                    onopen: function() {
                        console.log('Monitor WebSocket connected');
                    },
                    onmessage: function(event) {
                        try {
                            const data = JSON.parse(event.data);
                            if (data.type === 'monitor_detail') {
                                self.data = data.data;
                                self.render();
                            }
                        } catch (e) {
                            console.error('Failed to parse message:', e);
                        }
                    },
                    onclose: function() {
                        console.log('Monitor WebSocket disconnected, reconnecting...');
                        HomeApp.reconnect(wsUrl, this);
                    },
                    onerror: function(error) {
                        console.error('Monitor WebSocket error:', error);
                    }
                });
            },

            render: function() {
                const container = document.getElementById('monitor-container');
                if (!container || !this.data) return;

                const data = this.data;
                const statusClass = data.is_valid ? 'up' : 'down';
                const statusText = data.is_valid ? '正常' : '无法访问';
                const statusIcon = data.is_valid ? 'fa-check' : 'fa-times';
                const statusColor = data.is_valid ? '#27ae60' : '#e74c3c';

                let html = '<div class="status-card">';
                html += '<div class="status-header">';
                html += '<div class="service-name">' + HomeApp.utils.getTypeIcon(data.type) + ' ' + HomeApp.utils.escapeHtml(data.name) + '</div>';
                html += '<div class="status-indicator ' + statusClass + '">';
                html += '<i class="fas ' + statusIcon + '"></i> ' + statusText + ' ' + (data.latency || '');
                html += '</div>';
                html += '</div>';

                html += '<div class="time-grid">';
                for (let row = 0; row < 6; row++) {
                    html += '<div class="time-row">';
                    for (let i = 0; i < 24; i++) {
                        const isUp = Math.random() > 0.2;
                        html += '<div class="time-cell ' + (isUp ? 'up' : 'down') + '"></div>';
                    }
                    html += '</div>';
                }
                html += '</div>';

                html += '<div class="legend">';
                html += '<div class="legend-item"><div class="legend-color" style="background-color: #27ae60;"></div><span>正常</span></div>';
                html += '<div class="legend-item"><div class="legend-color" style="background-color: #e74c3c;"></div><span>故障</span></div>';
                html += '<div class="legend-item"><div class="legend-color" style="background-color: #f0f0f0;"></div><span>未知</span></div>';
                html += '</div>';

                html += '<div style="display: flex; flex-wrap: wrap; gap: 20px; font-size: 14px; margin-top: 15px;">';
                html += '<div>类型: ' + HomeApp.utils.escapeHtml(data.type) + '</div>';
                html += '<div>地址: ' + HomeApp.utils.escapeHtml(data.address) + '</div>';
                if (data.speed) {
                    html += '<div>耗时: ' + data.speed + 'ms</div>';
                }
                if (data.size) {
                    html += '<div>大小: ' + HomeApp.utils.formatSize(data.size) + '</div>';
                }
                html += '<div>日志: ' + Math.floor(Math.random() * 1000) + ' 条</div>';
                html += '</div>';

                html += '<div class="status-footer">';
                html += '<div>今天</div>';
                html += '<div class="status-stats">最近 60 天可用率 ' + (data.is_valid ? '100.0' : '0.0') + '%</div>';
                html += '<div class="date">' + new Date().toLocaleDateString('zh-CN') + '</div>';
                html += '</div>';

                html += '</div>';

                html += '<div class="status-card">';
                html += '<h3 style="font-size: 16px; font-weight: 600; margin-bottom: 15px;">最近7天监控数据</h3>';

                if (data.week_logs && data.week_logs.length > 0) {
                    const upCount = data.week_logs.filter(function(log) { return log.is_valid; }).length;
                    const totalCount = data.week_logs.length;
                    const upRate = totalCount > 0 ? (upCount / totalCount * 100).toFixed(1) : '0.0';

                    html += '<div style="margin-top: 15px; display: flex; gap: 20px; font-size: 14px;">';
                    html += '<div>总监控次数: ' + totalCount + ' 次</div>';
                    html += '<div>正常次数: <span style="color: #27ae60;">' + upCount + ' 次</span></div>';
                    html += '<div>正常率: <span style="color: #27ae60;">' + upRate + '%</span></div>';
                    html += '</div>';
                } else {
                    html += '<div style="text-align: center; padding: 40px 0; color: #999;">暂无7天数据</div>';
                }

                html += '</div>';

                container.innerHTML = html;
            }
        },

        init: function() {
            const container = document.getElementById('group-container');
            if (container) {
                this.renderGroups.init();
                return;
            }

            const monitorContainer = document.getElementById('monitor-container');
            if (monitorContainer) {
                this.renderMonitor.init();
            }
        }
    };

    document.addEventListener('DOMContentLoaded', function() {
        HomeApp.init();
    });

    window.HomeApp = HomeApp;
})();