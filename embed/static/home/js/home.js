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
                const wsUrl = HomeApp.utils.getWsUrl('/ws/groups');

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

                let html = '';
                this.data.forEach(function(group) {
                    const monitors = group.monitors || [];
                    const upCount = monitors.filter(function(m) { return m.is_valid; }).length;
                    const totalCount = monitors.length;
                    const upRate = totalCount > 0 ? (upCount / totalCount * 100).toFixed(1) : '0.0';

                    html += '<div class="status-card" data-group-id="' + group.id + '">';
                    html += '<div class="status-header">';
                    html += '<div class="service-name">📁 ' + HomeApp.utils.escapeHtml(group.name) + '</div>';
                    html += '<div class="status-indicator">';
                    html += '<span>正常: ' + upCount + '/' + totalCount + ' (' + upRate + '%)</span>';
                    html += '</div>';
                    html += '</div>';

                    if (monitors.length > 0) {
                        html += '<div style="margin-top: 15px;">';
                        monitors.forEach(function(monitor) {
                            const statusClass = monitor.is_valid ? 'up' : 'down';
                            const statusText = monitor.is_valid ? '正常' : '无法访问';
                            const statusIcon = monitor.is_valid ? 'fa-check' : 'fa-times';

                            html += '<div style="display: flex; align-items: center; margin-bottom: 8px; padding: 8px; background-color: #f8f9fa; border-radius: 4px;">';
                            html += '<div style="margin-right: 10px;">' + HomeApp.utils.getTypeIcon(monitor.type) + '</div>';
                            html += '<div style="flex: 1;">';
                            html += '<div style="font-size: 14px; font-weight: 500;">' + HomeApp.utils.escapeHtml(monitor.name) + '</div>';
                            html += '<div style="font-size: 12px; color: #666;">' + HomeApp.utils.escapeHtml(monitor.address) + '</div>';
                            html += '</div>';
                            html += '<div class="status-indicator ' + statusClass + '">';
                            html += '<i class="fas ' + statusIcon + '"></i>';
                            html += '<span>' + statusText + '</span>';
                            if (monitor.latency) {
                                html += '<span style="margin-left: 5px; font-size: 12px;">' + HomeApp.utils.escapeHtml(monitor.latency) + '</span>';
                            }
                            html += '</div>';
                            html += '</div>';
                        });
                        html += '</div>';
                    } else {
                        html += '<div style="text-align: center; padding: 20px; color: #999; font-size: 14px;">该分组下暂无监控点</div>';
                    }

                    html += '<div class="status-footer">';
                    html += '<div class="status-stats">今天</div>';
                    html += '<div class="status-stats">最近 60 天可用率 ' + upRate + '%</div>';
                    html += '<div class="date">' + new Date().toLocaleDateString('zh-CN') + '</div>';
                    html += '</div>';
                    html += '</div>';
                }.bind(this));

                container.innerHTML = html;
            },

            renderTabs: function() {
                const tabsContainer = document.getElementById('group-tabs');
                if (!tabsContainer) return;

                let html = '<div class="tab active" data-tab="all">全部分组</div>';

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
                const groupCards = document.querySelectorAll('.status-card');

                tabs.forEach(function(tab) {
                    tab.onclick = function() {
                        tabs.forEach(function(t) { t.classList.remove('active'); });
                        tab.classList.add('active');

                        const tabName = tab.dataset.tab;

                        groupCards.forEach(function(card) {
                            if (tabName === 'all' || card.dataset.groupId === tabName.replace('group-', '')) {
                                card.style.display = 'block';
                            } else {
                                card.style.display = 'none';
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

                let html = '<div class="status-card">';
                html += '<div class="status-header">';
                html += '<div class="service-name">' + HomeApp.utils.getTypeIcon(data.type) + ' ' + HomeApp.utils.escapeHtml(data.name) + '</div>';
                html += '<div class="status-indicator ' + statusClass + '">';
                html += '<i class="fas ' + statusIcon + '"></i>';
                html += '<span>' + statusText + '</span>';
                if (data.latency) {
                    html += '<span class="latency">' + HomeApp.utils.escapeHtml(data.latency) + '</span>';
                }
                html += '</div>';
                html += '</div>';

                html += '<div style="margin-top: 20px;">';
                html += '<div style="display: grid; grid-template-columns: 120px 1fr; gap: 10px; margin-bottom: 20px;">';
                html += '<div style="font-weight: 500;">地址:</div>';
                html += '<div>' + HomeApp.utils.escapeHtml(data.address) + '</div>';
                html += '<div style="font-weight: 500;">类型:</div>';
                html += '<div>' + HomeApp.utils.escapeHtml(data.type) + '</div>';
                html += '<div style="font-weight: 500;">分组:</div>';
                html += '<div>' + (data.group_name ? HomeApp.utils.escapeHtml(data.group_name) : '未分组') + '</div>';
                html += '<div style="font-weight: 500;">创建时间:</div>';
                html += '<div>' + HomeApp.utils.formatDate(data.created_at) + '</div>';
                html += '<div style="font-weight: 500;">更新时间:</div>';
                html += '<div>' + HomeApp.utils.formatDate(data.updated_at) + '</div>';
                html += '</div>';
                html += '</div>';

                html += '<div style="margin-top: 30px;">';
                html += '<h3 style="font-size: 16px; font-weight: 600; margin-bottom: 15px;">最近7天监控数据</h3>';

                if (data.week_logs && data.week_logs.length > 0) {
                    html += '<div class="time-grid">';
                    data.week_logs.forEach(function(log) {
                        const status = log.is_valid ? 'up' : 'down';
                        const error = log.error_msg || '无法访问';
                        const dateStr = log.date || '';
                        let cellHtml = '<div class="time-cell ' + status + '" data-status="' + status + '" data-time="' + HomeApp.utils.escapeHtml(dateStr) + '" data-error="' + HomeApp.utils.escapeHtml(error) + '"';
                        if (log.speed) {
                            cellHtml += ' data-speed="' + log.speed + '"';
                        }
                        if (log.size !== undefined && log.size !== null) {
                            cellHtml += ' data-size="' + HomeApp.utils.formatSize(log.size) + '"';
                        }
                        cellHtml += '></div>';
                        html += cellHtml;
                    });
                    html += '</div>';

                    const upCount = data.week_logs.filter(function(log) { return log.is_valid; }).length;
                    const totalCount = data.week_logs.length;
                    const upRate = totalCount > 0 ? (upCount / totalCount * 100).toFixed(1) : '0.0';

                    html += '<div style="margin-top: 15px; display: flex; gap: 20px; font-size: 14px;">';
                    html += '<div>总监控次数: ' + totalCount + ' 次</div>';
                    html += '<div>正常次数: <span style="color: #27ae60;">' + upCount + ' 次</span></div>';
                    html += '<div>正常率: <span style="color: #27ae60;">' + upRate + '%</span></div>';
                    html += '</div>';

                    html += '<div class="status-footer">';
                    html += '<div class="status-stats">今天</div>';
                    html += '<div class="status-stats">最近 60 天可用率 ' + upRate + '%</div>';
                    html += '<div class="date">' + new Date().toLocaleDateString('zh-CN') + '</div>';
                    html += '</div>';
                } else {
                    html += '<div style="text-align: center; padding: 40px 0; color: #999;">暂无7天数据</div>';
                }

                html += '</div>';
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
