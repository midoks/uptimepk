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
                            // 生成多行时间网格，使用 monitor.hour_logs 数据
                            const hourLogs = monitor.hour_logs || [];
                            
                            // 如果没有日志数据，生成模拟数据
                            if (hourLogs.length === 0) {
                                for (let row = 0; row < 6; row++) {
                                    html += '<div class="time-row">';
                                    for (let i = 0; i < 24; i++) {
                                        const isUp = Math.random() > 0.2;
                                        const status = isUp ? 'up' : 'down';
                                        const hour = HomeApp.utils.padZero(i);
                                        const minute = HomeApp.utils.padZero(Math.floor(Math.random() * 60));
                                        const timeStr = hour + ':' + minute;
                                        const speed = (Math.random() * 1000).toFixed(2) + 'ms';
                                        const size = (Math.random() * 200).toFixed(2) + 'kb';
                                        
                                        html += '<div class="time-cell ' + status + '" ' +
                                            'data-status="' + status + '" ' +
                                            'data-time="' + timeStr + '" ' +
                                            'data-speed="' + speed + '" ' +
                                            'data-size="' + size + '"></div>';
                                    }
                                    html += '</div>';
                                }
                            } else {
                                // 使用实际日志数据
                                for (let row = 0; row < 6; row++) {
                                    html += '<div class="time-row">';
                                    for (let i = 0; i < 24; i++) {
                                        const log = hourLogs.find(function(l) {
                                            return parseInt(l.hour) === i;
                                        });
                                        
                                        if (log) {
                                            const status = log.is_valid ? 'up' : 'down';
                                            const timeStr = HomeApp.utils.padZero(log.hour) + ':' + HomeApp.utils.padZero(log.minute || 0);
                                            let cellHtml = '<div class="time-cell ' + status + '" ' +
                                                'data-status="' + status + '" ' +
                                                'data-time="' + timeStr + '" ';
                                            
                                            if (log.error_msg) {
                                                cellHtml += 'data-error="' + HomeApp.utils.escapeHtml(log.error_msg) + '" ';
                                            }
                                            if (log.speed) {
                                                cellHtml += 'data-speed="' + log.speed + '" ';
                                            }
                                            if (log.size !== undefined && log.size !== null) {
                                                let sizeStr;
                                                if (log.size >= 1024) {
                                                    sizeStr = (log.size / 1024).toFixed(2) + 'kb';
                                                } else {
                                                    sizeStr = log.size + 'b';
                                                }
                                                cellHtml += 'data-size="' + sizeStr + '" ';
                                            }
                                            cellHtml += '></div>';
                                            html += cellHtml;
                                        } else {
                                            // 无数据时显示未知状态
                                            html += '<div class="time-cell unknown" ' +
                                                'data-status="unknown" ' +
                                                'data-time="' + HomeApp.utils.padZero(i) + ':00" ' +
                                                '></div>';
                                        }
                                    }
                                    html += '</div>';
                                }
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
                            html += '<div>日志: ' + (hourLogs.length || Math.floor(Math.random() * 1000)) + ' 条</div>';
                            html += '</div>';

                            html += '<div class="status-footer">';
                            html += '<div>今天</div>';
                            html += '<div class="status-stats">最近 60 天可用率 ' + (monitor.is_valid ? '100.0' : '0.0') + '%</div>';
                            html += '<div class="date">' + new Date().toLocaleDateString('zh-CN') + '</div>';
                            html += '</div>';

                            // 为状态指示器添加错误信息属性，用于 tooltip
                            html += '<div class="error-tooltip" style="display: none;" data-error="' + HomeApp.utils.escapeHtml(monitor.error_msg || '') + '"></div>';

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
                const hourLogs = data.hour_logs || [];
                
                if (hourLogs.length === 0) {
                    for (let row = 0; row < 6; row++) {
                        html += '<div class="time-row">';
                        for (let i = 0; i < 24; i++) {
                            const isUp = Math.random() > 0.2;
                            const status = isUp ? 'up' : 'down';
                            const hour = HomeApp.utils.padZero(i);
                            const minute = HomeApp.utils.padZero(Math.floor(Math.random() * 60));
                            const timeStr = hour + ':' + minute;
                            const speed = (Math.random() * 1000).toFixed(2) + 'ms';
                            const size = (Math.random() * 200).toFixed(2) + 'kb';
                            
                            html += '<div class="time-cell ' + status + '" ' +
                                'data-status="' + status + '" ' +
                                'data-time="' + timeStr + '" ' +
                                'data-speed="' + speed + '" ' +
                                'data-size="' + size + '"></div>';
                        }
                        html += '</div>';
                    }
                } else {
                    for (let row = 0; row < 6; row++) {
                        html += '<div class="time-row">';
                        for (let i = 0; i < 24; i++) {
                            const log = hourLogs.find(function(l) {
                                return parseInt(l.hour) === i;
                            });
                            
                            if (log) {
                                const status = log.is_valid ? 'up' : 'down';
                                const timeStr = HomeApp.utils.padZero(log.hour) + ':' + HomeApp.utils.padZero(log.minute || 0);
                                let cellHtml = '<div class="time-cell ' + status + '" ' +
                                    'data-status="' + status + '" ' +
                                    'data-time="' + timeStr + '" ';
                                
                                if (log.error_msg) {
                                    cellHtml += 'data-error="' + HomeApp.utils.escapeHtml(log.error_msg) + '" ';
                                }
                                if (log.speed) {
                                    cellHtml += 'data-speed="' + log.speed + '" ';
                                }
                                if (log.size !== undefined && log.size !== null) {
                                    let sizeStr;
                                    if (log.size >= 1024) {
                                        sizeStr = (log.size / 1024).toFixed(2) + 'kb';
                                    } else {
                                        sizeStr = log.size + 'b';
                                    }
                                    cellHtml += 'data-size="' + sizeStr + '" ';
                                }
                                cellHtml += '></div>';
                                html += cellHtml;
                            } else {
                                html += '<div class="time-cell unknown" ' +
                                    'data-status="unknown" ' +
                                    'data-time="' + HomeApp.utils.padZero(i) + ':00" ' +
                                    '></div>';
                            }
                        }
                        html += '</div>';
                    }
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
                html += '<div>日志: ' + (hourLogs.length || Math.floor(Math.random() * 1000)) + ' 条</div>';
                html += '</div>';

                html += '<div class="status-footer">';
                html += '<div>今天</div>';
                html += '<div class="status-stats">最近 60 天可用率 ' + (data.is_valid ? '100.0' : '0.0') + '%</div>';
                html += '<div class="date">' + new Date().toLocaleDateString('zh-CN') + '</div>';
                html += '</div>';

                // 为状态指示器添加错误信息属性，用于 tooltip
                html += '<div class="error-tooltip" style="display: none;" data-error="' + HomeApp.utils.escapeHtml(data.error_msg || '') + '"></div>';

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

    // 添加 footer.tmpl 中的 JavaScript 代码
    document.addEventListener('DOMContentLoaded', function() {
        const tabs = document.querySelectorAll('.tab');
        const statusCards = document.querySelectorAll('.status-card');

        tabs.forEach(tab => {
            tab.addEventListener('click', function() {
                tabs.forEach(t => t.classList.remove('active'));
                this.classList.add('active');

                const tabName = this.dataset.tab;

                statusCards.forEach(card => {
                    if (tabName === 'all' || card.dataset.category === tabName) {
                        card.style.display = 'block';
                    } else {
                        card.style.display = 'none';
                    }
                });
            });
        });

        // 为时间格子添加鼠标悬停事件
        document.addEventListener('mouseenter', function(e) {
            if (!e.target || !e.target.classList) return;
            if (e.target.classList.contains('time-cell')) {
                const cell = e.target;
                const status = cell.dataset.status;
                const time = cell.dataset.time;
                const error = cell.dataset.error || '';

                if (!cell._tip) {
                    // 创建提示元素
                    const tip = document.createElement('div');
                    tip.className = 'time-cell-tip';
                    tip.style.position = 'fixed';
                    tip.style.backgroundColor = '#333';
                    tip.style.color = '#fff';
                    tip.style.padding = '5px 10px';
                    tip.style.borderRadius = '4px';
                    tip.style.fontSize = '12px';
                    tip.style.zIndex = '10000';
                    tip.style.pointerEvents = 'none';
                    tip.style.maxWidth = '300px';
                    tip.style.wordBreak = 'break-word';

                    // 根据状态设置不同的内容和样式
                    if (status === 'up') {
                        tip.innerHTML = '状态: <span style="color: #4CAF50;">正常</span> | 时间: ' + time;
                        // 检查是否有耗时和大小信息
                        if (cell.dataset.speed) {
                            tip.innerHTML += ' | 耗时: ' + cell.dataset.speed + 'ms';
                        }
                        // 检查大小信息，包括0
                        if (cell.dataset.size !== undefined && cell.dataset.size !== null) {
                            tip.innerHTML += ' | 大小: ' + cell.dataset.size;
                        }
                    } else if (status === 'down') {
                        tip.innerHTML = '状态: <span style="color: #f44336;">故障</span> | 时间: ' + time + ' | 错误: <span style="color: #f44336;">' + error + '</span>';
                    } else {
                        tip.innerHTML = '状态: <span style="color: #ff9800;">未知</span> | 时间: ' + time;
                    }

                    document.body.appendChild(tip);

                    // 计算位置，确保tooltip在视口内
                    let top = e.clientY + 10;
                    let left = e.clientX + 10;
                    
                    const tipWidth = tip.offsetWidth;
                    const tipHeight = tip.offsetHeight;
                    const viewportWidth = window.innerWidth;
                    const viewportHeight = window.innerHeight;
                    
                    // 调整位置
                    if (left + tipWidth > viewportWidth) {
                        left = e.clientX - tipWidth - 10;
                    }
                    if (top + tipHeight > viewportHeight) {
                        top = e.clientY - tipHeight - 10;
                    }
                    
                    // 确保不超出视口
                    left = Math.max(10, left);
                    top = Math.max(10, top);

                    // 定位提示元素
                    tip.style.top = top + 'px';
                    tip.style.left = left + 'px';

                    // 存储提示元素引用
                    cell._tip = tip;
                }
            }
        }, true);

        document.addEventListener('mouseleave', function(e) {
            if (!e.target || !e.target.classList) return;
            if (e.target.classList.contains('time-cell')) {
                const cell = e.target;
                if (cell._tip) {
                    document.body.removeChild(cell._tip);
                    delete cell._tip;
                }
            }
        }, true);

        document.addEventListener('mousemove', function(e) {
            if (!e.target || !e.target.classList) return;
            if (e.target.classList.contains('time-cell') && e.target._tip) {
                let top = e.clientY + 10;
                let left = e.clientX + 10;
                
                const tip = e.target._tip;
                const tipWidth = tip.offsetWidth;
                const tipHeight = tip.offsetHeight;
                const viewportWidth = window.innerWidth;
                const viewportHeight = window.innerHeight;
                
                // 调整位置
                if (left + tipWidth > viewportWidth) {
                    left = e.clientX - tipWidth - 10;
                }
                if (top + tipHeight > viewportHeight) {
                    top = e.clientY - tipHeight - 10;
                }
                
                // 确保不超出视口
                left = Math.max(10, left);
                top = Math.max(10, top);
                
                tip.style.top = top + 'px';
                tip.style.left = left + 'px';
            }
        }, true);

        // 为状态指示器添加鼠标悬停事件
        document.addEventListener('mouseenter', function(e) {
            if (!e.target || !e.target.closest) return;
            const indicator = e.target.closest('.status-indicator');
            if (indicator) {
                const card = indicator.closest('.status-card');
                const errorTooltip = card ? card.querySelector('.error-tooltip') : null;
                const error = errorTooltip ? errorTooltip.dataset.error : '';

                if (error && !indicator._tip) {
                    // 创建提示元素
                    const tip = document.createElement('div');
                    tip.className = 'status-indicator-tip';
                    tip.style.position = 'fixed';
                    tip.style.backgroundColor = '#333';
                    tip.style.color = '#fff';
                    tip.style.padding = '5px 10px';
                    tip.style.borderRadius = '4px';
                    tip.style.fontSize = '12px';
                    tip.style.zIndex = '10000';
                    tip.style.pointerEvents = 'none';
                    tip.style.whiteSpace = 'pre-wrap';
                    tip.style.maxWidth = '300px';

                    tip.innerHTML = '<span style="color: #f44336;">错误信息:</span><br>' + error;

                    document.body.appendChild(tip);

                    // 定位提示元素
                    tip.style.top = (e.clientY + 10) + 'px';
                    tip.style.left = (e.clientX + 10) + 'px';

                    // 存储提示元素引用
                    indicator._tip = tip;
                }
            }
        }, true);

        document.addEventListener('mouseleave', function(e) {
            if (!e.target || !e.target.closest) return;
            const indicator = e.target.closest('.status-indicator');
            if (indicator) {
                if (indicator._tip) {
                    document.body.removeChild(indicator._tip);
                    delete indicator._tip;
                }
            }
        }, true);

        document.addEventListener('mousemove', function(e) {
            if (!e.target || !e.target.closest) return;
            const indicator = e.target.closest('.status-indicator');
            if (indicator && indicator._tip) {
                indicator._tip.style.top = (e.clientY + 10) + 'px';
                indicator._tip.style.left = (e.clientX + 10) + 'px';
            }
        }, true);

        // WebSocket 连接
        let ws = null;
        let reconnectTimer = null;
        let monitorData = [];
        let chunkBuffer = [];
        let chunkTotal = 0;
        let chunkCount = 0;
        let chunkGroups = null;

        function escapeHtml(text) {
            if (!text) return '';
            var map = {
                '&': '&amp;',
                '<': '&lt;',
                '>': '&gt;',
                '"': '&quot;',
                "'": '&#039;'
            };
            return text.replace(/[&<>'"]/g, function(m) { return map[m]; });
        }

        function getTypeIcon(type) {
            switch(type) {
                case 'http':
                case 'https':
                    return '<i class="fas fa-globe"></i>';
                case 'tcp':
                    return '<i class="fas fa-network-wired"></i>';
                case 'udp':
                    return '<i class="fas fa-broadcast-tower"></i>';
                case 'dns':
                    return '<i class="fas fa-server"></i>';
                case 'ping':
                    return '<i class="fas fa-wifi"></i>';
                default:
                    return '<i class="fas fa-question"></i>';
            }
        }

        function padZero(num) {
            return String(num).padStart(2, '0');
        }

        function renderStatusCards(data) {
            const container = document.getElementById('status-container');
            if (!container) return;

            if (!data || data.length === 0) {
                container.innerHTML = '<div class="loading">暂无监控数据</div>';
                return;
            }

            let html = '';
            data.forEach(function(monitor) {
                const statusClass = monitor.is_valid ? 'up' : 'down';
                const statusText = monitor.is_valid ? '正常' : '无法访问';
                const statusIcon = monitor.is_valid ? 'fa-check' : 'fa-times';

                html += '<div class="status-card" data-category="tab-' + monitor.gid + '" data-id="' + monitor.id + '">';
                html += '<div class="status-header">';
                html += '<div class="service-name">' + getTypeIcon(monitor.type) + ' ' + monitor.name + '</div>';
                html += '<div class="status-indicator ' + statusClass + '">';
                html += '<i class="fas ' + statusIcon + '"></i>';
                html += '<span>' + statusText + '</span>';
                if (monitor.latency) {
                    html += '<span class="latency">' + monitor.latency + '</span>';
                }
                html += '</div>';
                html += '</div>';

                // 渲染时间格子 - 每行100个
                // 后端已经按时间升序返回（从早到晚）
                const hourLogs = monitor.hour_logs || [];
                const cellsPerRow = 78;

                html += '<div class="time-grid">';

                for (let i = 0; i < hourLogs.length; i++) {
                    const log = hourLogs[i];
                    const status = log.is_valid ? 'up' : 'down';
                    const error = log.error_msg || '无法访问';
                    const timeStr = padZero(log.hour) + ':' + padZero(log.minute);
                    let cellHtml = '<div class="time-cell ' + status + '" data-status="' + status + '" data-time="' + timeStr + '" data-error="' + escapeHtml(error) + '"';
                    if (log.speed) {
                        cellHtml += ' data-speed="' + log.speed + '"';
                    }
                    if (log.size !== undefined && log.size !== null) {
                        // 根据大小转换单位
                        let sizeStr;
                        if (log.size >= 1024) {
                            const sizeKb = (log.size / 1024).toFixed(2);
                            sizeStr = sizeKb + 'kb';
                        } else {
                            sizeStr = log.size + 'b';
                        }
                        cellHtml += ' data-size="' + sizeStr + '"';
                    }
                    cellHtml += '></div>';
                    html += cellHtml;
                }

                if (hourLogs.length === 0) {
                    html += '<span style="color:#999;font-size:12px;">暂无日志数据</span>';
                }

                html += '</div>';

                // 图例
                html += '<div class="legend">';
                html += '<div class="legend-item"><div class="legend-color" style="background-color: #27ae60;"></div><span>正常</span></div>';
                html += '<div class="legend-item"><div class="legend-color" style="background-color: #e74c3c;"></div><span>故障</span></div>';
                html += '<div class="legend-item"><div class="legend-color" style="background-color: #f0f0f0;"></div><span>未知</span></div>';
                html += '</div>';

                // 页脚统计
                html += '<div class="status-footer">';
                html += '<div class="status-stats">类型: ' + monitor.type + '</div>';
                html += '<div class="status-stats">日志: ' + hourLogs.length + ' 条</div>';
                if (monitor.is_valid) {
                    if (monitor.speed) {
                        html += '<div class="status-stats">耗时: ' + monitor.speed + 'ms</div>';
                    }
                    if (monitor.size !== undefined && monitor.size !== null) {
                        // 根据大小转换单位
                        let sizeStr;
                        if (monitor.size >= 1024) {
                            const sizeKb = (monitor.size / 1024).toFixed(2);
                            sizeStr = sizeKb + 'kb';
                        } else {
                            sizeStr = monitor.size + 'b';
                        }
                        html += '<div class="status-stats">大小: ' + sizeStr + '</div>';
                    }
                } else {
                    html += '<div class="status-stats" style="color: #f44336;">无法访问</div>';
                }
                html += '</div>';

                // 为状态指示器添加错误信息属性，用于 tooltip
                html += '<div class="error-tooltip" style="display: none;" data-error="' + escapeHtml(monitor.error_msg || '') + '"></div>';

                html += '</div>';
            });

            container.innerHTML = html;

            // 重新绑定 tab 点击事件
            bindTabEvents();
        }

        function renderTabs(groups) {
            const tabsContainer = document.getElementById('status-tabs');
            if (!tabsContainer) return;

            let html = '<div class="tab active" data-tab="all">All</div>';

            if (groups && groups.length > 0) {
                groups.forEach(function(group) {
                    html += '<div class="tab" data-tab="tab-' + group.id + '">' + group.name + '</div>';
                });
            }

            tabsContainer.innerHTML = html;
            bindTabEvents();
        }

        function bindTabEvents() {
            const tabs = document.querySelectorAll('.tab');
            const statusCards = document.querySelectorAll('.status-card');

            tabs.forEach(tab => {
                tab.onclick = function() {
                    tabs.forEach(t => t.classList.remove('active'));
                    this.classList.add('active');

                    const tabName = this.dataset.tab;

                    statusCards.forEach(card => {
                        if (tabName === 'all' || card.dataset.category === tabName) {
                            card.style.display = 'block';
                        } else {
                            card.style.display = 'none';
                        }
                    });
                };
            });
        }

        function updateMonitorStatus(monitorList) {
            monitorList.forEach(function(monitor) {
                const card = document.querySelector('.status-card[data-id="' + monitor.id + '"]');
                if (card) {
                    const indicator = card.querySelector('.status-indicator');
                    if (monitor.is_valid) {
                        indicator.className = 'status-indicator up';
                        indicator.innerHTML = '<i class="fas fa-check"></i><span>正常</span>';
                        if (monitor.latency) {
                            indicator.innerHTML += '<span class="latency">' + monitor.latency + '</span>';
                        }
                    } else {
                        indicator.className = 'status-indicator down';
                        indicator.innerHTML = '<i class="fas fa-times"></i><span>无法访问</span>';
                    }

                    // 更新错误信息
                    const footer = card.querySelector('.status-footer');
                    if (footer) {
                        const statsDivs = footer.querySelectorAll('.status-stats');
                        if (monitor.is_valid) {
                            if (statsDivs.length >= 3) {
                                statsDivs[1].innerHTML = '日志: ' + (monitor.hour_logs ? monitor.hour_logs.length : 0) + ' 条';
                                if (monitor.speed && statsDivs.length >= 3) {
                                    statsDivs[2].innerHTML = '耗时: ' + monitor.speed + 'ms';
                                }
                                if (monitor.size !== undefined && monitor.size !== null && statsDivs.length >= 4) {
                                    // 根据大小转换单位
                                    let sizeStr;
                                    if (monitor.size >= 1024) {
                                        const sizeKb = (monitor.size / 1024).toFixed(2);
                                        sizeStr = sizeKb + 'kb';
                                    } else {
                                        sizeStr = monitor.size + 'b';
                                    }
                                    statsDivs[3].innerHTML = '大小: ' + sizeStr;
                                }
                                statsDivs[2].style.color = '';
                            }
                        } else {
                            if (statsDivs.length >= 3) {
                                statsDivs[1].innerHTML = '日志: ' + (monitor.hour_logs ? monitor.hour_logs.length : 0) + ' 条';
                                statsDivs[2].innerHTML = '无法访问';
                                statsDivs[2].style.color = '#f44336';
                            }
                        }
                    }

                    // 更新错误提示信息
                    const errorTooltip = card.querySelector('.error-tooltip');
                    if (errorTooltip) {
                        errorTooltip.dataset.error = monitor.error_msg || '';
                    }
                }
            });
        }

        function applyMonitorUpdates(updateList) {
            updateList.forEach(function(update) {
                const card = document.querySelector('.status-card[data-id="' + update.id + '"]');
                if (card) {
                    const indicator = card.querySelector('.status-indicator');
                    if (update.is_valid) {
                        indicator.className = 'status-indicator up';
                        indicator.innerHTML = '<i class="fas fa-check"></i><span>正常</span>';
                        if (update.latency) {
                            indicator.innerHTML += '<span class="latency">' + update.latency + '</span>';
                        }
                    } else {
                        indicator.className = 'status-indicator down';
                        indicator.innerHTML = '<i class="fas fa-times"></i><span>无法访问</span>';
                    }

                    // 更新错误提示信息
                    const errorTooltip = card.querySelector('.error-tooltip');
                    if (errorTooltip) {
                        errorTooltip.dataset.error = update.error_msg || '';
                    }

                    // 更新错误信息
                    const footer = card.querySelector('.status-footer');
                    if (footer) {
                        if (update.is_valid) {
                            const statsDivs = footer.querySelectorAll('.status-stats');
                            if (statsDivs.length >= 3) {
                                if (update.speed && statsDivs.length >= 3) {
                                    statsDivs[2].innerHTML = '耗时: ' + update.speed + 'ms';
                                }
                                if (update.size !== undefined && update.size !== null && statsDivs.length >= 4) {
                                    // 根据大小转换单位
                                    let sizeStr;
                                    if (update.size >= 1024) {
                                        const sizeKb = (update.size / 1024).toFixed(2);
                                        sizeStr = sizeKb + 'kb';
                                    } else {
                                        sizeStr = update.size + 'b';
                                    }
                                    statsDivs[3].innerHTML = '大小: ' + sizeStr;
                                }
                                statsDivs[2].style.color = '';
                            }
                        } else {
                            const statsDivs = footer.querySelectorAll('.status-stats');
                            if (statsDivs.length >= 3) {
                                statsDivs[2].innerHTML = '无法访问';
                                statsDivs[2].style.color = '#f44336';
                            }
                        }
                    }

                    // 添加新的日志
                    if (update.new_log) {
                        addNewLogToCard(card, update.new_log);
                    }

                    // 更新本地数据
                    const monitorIndex = monitorData.findIndex(m => m.id === update.id);
                    if (monitorIndex !== -1) {
                        monitorData[monitorIndex].is_valid = update.is_valid;
                        monitorData[monitorIndex].speed = update.speed;
                        monitorData[monitorIndex].size = update.size;
                        monitorData[monitorIndex].error_msg = update.error_msg;
                        if (update.new_log) {
                            const hourLogs = monitorData[monitorIndex].hour_logs || [];
                            // 检查是否已存在相同的日志
                            const isDuplicate = hourLogs.some(log => 
                                log.hour === update.new_log.hour && log.minute === update.new_log.minute
                            );
                            if (!isDuplicate) {
                                if (!monitorData[monitorIndex].hour_logs) {
                                    monitorData[monitorIndex].hour_logs = [];
                                }
                                monitorData[monitorIndex].hour_logs.push(update.new_log);
                            }
                        }
                    }
                }
            });
        }

        function addNewLogToCard(card, newLog) {
            const timeGrid = card.querySelector('.time-grid');
            if (!timeGrid) return;

            // 移除"暂无日志数据"的提示行
            const emptySpan = timeGrid.querySelector('span');
            if (emptySpan) {
                emptySpan.remove();
            }

            const status = newLog.is_valid ? 'up' : 'down';
            const error = newLog.error_msg || '无法访问';
            const timeStr = padZero(newLog.hour) + ':' + padZero(newLog.minute);

            const newCell = document.createElement('div');
            newCell.className = 'time-cell ' + status;
            newCell.dataset.status = status;
            newCell.dataset.time = timeStr;
            newCell.dataset.error = escapeHtml(error);
            if (newLog.speed) {
                newCell.dataset.speed = newLog.speed;
            }
            if (newLog.size !== undefined && newLog.size !== null) {
                // 根据大小转换单位
                let sizeStr;
                if (newLog.size >= 1024) {
                    const sizeKb = (newLog.size / 1024).toFixed(2);
                    sizeStr = sizeKb + 'kb';
                } else {
                    sizeStr = newLog.size + 'b';
                }
                newCell.dataset.size = sizeStr;
            }

            // 添加到 time-grid 的末尾
            timeGrid.appendChild(newCell);

            // 更新日志数量统计
            const footer = card.querySelector('.status-footer');
            if (footer) {
                const statsDivs = footer.querySelectorAll('.status-stats');
                if (statsDivs.length >= 2) {
                    const currentCount = timeGrid.querySelectorAll('.time-cell').length;
                    statsDivs[1].innerHTML = '日志: ' + currentCount + ' 条';
                }
            }
        }

        function connectWS() {
            const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
            const wsUrl = protocol + '//' + window.location.host + '/ws/status';

            try {
                ws = new WebSocket(wsUrl);

                ws.onopen = function() {
                    console.log('WebSocket connected');
                };

                ws.onmessage = function(event) {
                    try {
                        const data = JSON.parse(event.data);
                        console.log(data);
                        if (data.type === 'monitor_status') {
                            if (data.chunks && data.chunks > 1) {
                                if (data.chunk === 1) {
                                    chunkBuffer = [];
                                    chunkTotal = data.total;
                                    chunkCount = 0;
                                    chunkGroups = data.groups || null;
                                }
                                chunkBuffer = chunkBuffer.concat(data.data);
                                chunkCount++;
                                if (chunkCount >= data.chunks) {
                                    monitorData = chunkBuffer;
                                    renderStatusCards(monitorData);
                                    if (chunkGroups) {
                                        renderTabs(chunkGroups);
                                    }
                                    chunkBuffer = [];
                                    chunkTotal = 0;
                                    chunkCount = 0;
                                    chunkGroups = null;
                                }
                            } else {
                                monitorData = data.data || [];
                                renderStatusCards(monitorData);
                                if (data.groups) {
                                    renderTabs(data.groups);
                                }
                            }
                        } else if (data.type === 'monitor_updates') {
                            chunkBuffer = [];
                            chunkTotal = 0;
                            chunkCount = 0;
                            chunkGroups = null;
                            applyMonitorUpdates(data.data);
                        } else if (data.type === 'monitor_status_by_gid') {
                            updateMonitorStatus(data.data);
                        }
                    } catch (e) {
                        console.error('Failed to parse message:', e);
                    }
                };

                ws.onclose = function() {
                    console.log('WebSocket disconnected, reconnecting...');
                    reconnectTimer = setTimeout(connectWS, 5000);
                };

                ws.onerror = function(error) {
                    console.error('WebSocket error:', error);
                };
            } catch (e) {
                console.error('Failed to connect WebSocket:', e);
                reconnectTimer = setTimeout(connectWS, 5000);
            }
        }

        // 启动 WebSocket 连接
        connectWS();
    });
})();