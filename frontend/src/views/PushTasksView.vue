<template>
  <div class="tab-content">
    <!-- 错误提示 -->
    <div v-if="showError && errorMessage" class="error-message">
      {{ errorMessage }}
    </div>

    <h3>Push Tasks</h3>

    <!-- 如果没有数据，显示提示信息 -->
    <div v-if="tasks.length === 0" class="no-data-message">
      <p>暂无任务数据</p>
      <p v-if="errorMessage" class="error-hint">{{ errorMessage }}</p>
    </div>

    <!-- API调试面板 -->
    <div v-if="SHOW_API_DIAGNOSTICS" class="api-debug-panel">
      <h4>API路径问题诊断工具</h4>
      <p class="info-text">根据日志，只有 <code>/api/promqls</code> 能正常工作，其他API路径都返回HTML。这表明后端可能使用不同的路由前缀。</p>
      
      <div class="api-test-controls">
        <div class="api-path-options">
    <div>
            <input type="radio" id="path-option-1" name="api-path" value="/api" v-model="apiPathOption" @change="changeApiPath">
            <label for="path-option-1">使用 <code>/api</code> 前缀 (当前)</label>
    </div>
    <div>
            <input type="radio" id="path-option-2" name="api-path" value="" v-model="apiPathOption" @change="changeApiPath">
            <label for="path-option-2">无前缀 (直接使用 <code>/metrics_sources</code>)</label>
    </div>
    <div>
            <input type="radio" id="path-option-3" name="api-path" value="/api/v1" v-model="apiPathOption" @change="changeApiPath">
            <label for="path-option-3">使用 <code>/api/v1</code> 前缀</label>
    </div>
    <div>
            <input type="radio" id="path-option-custom" name="api-path" value="custom" @change="selectCustomPath">
            <label for="path-option-custom">自定义路径前缀:</label>
            <input type="text" v-model="customApiPath" placeholder="/custom/path" 
                  :disabled="apiPathOption !== 'custom'" @input="updateCustomPath">
          </div>
        </div>
        
        <div class="api-test-actions">
          <button class="test-api-btn" @click="testApiConnection">测试API连接</button>
          <button class="view-logs-btn" @click="showApiLogs = !showApiLogs">
            {{ showApiLogs ? '隐藏日志' : '查看API日志' }}
          </button>
          <button class="use-mock-btn" @click="enableMockData">使用模拟数据</button>
        </div>
      </div>
      
      <!-- API测试结果 -->
      <div v-if="apiTestResults.length > 0" class="api-test-results">
        <h5>API测试结果:</h5>
        <div v-for="(result, index) in apiTestResults" :key="index" class="api-test-result">
          <div :class="['result-badge', result.success ? 'success' : 'error']">
            {{ result.success ? '成功' : '失败' }}
          </div>
          <div class="result-endpoint">{{ result.endpoint }}</div>
          <div class="result-message">{{ result.message }}</div>
        </div>
        
        <div v-if="recommendedApiPath" class="api-recommendation">
          <h5>发现有效的API路径模式!</h5>
          <div class="recommendation-content">
            <p>推荐使用: <code>{{ recommendedApiPath }}</code></p>
            <button @click="applyRecommendedPath" class="apply-recommendation-btn">应用推荐配置</button>
          </div>
        </div>
      </div>
      
      <!-- 自动检测按钮 -->
      <div class="auto-detect-section">
        <button @click="autoDetectApiPath" class="auto-detect-btn" :disabled="isAutoDetecting">
          {{ isAutoDetecting ? '检测中...' : '自动检测API路径' }}
        </button>
        <span v-if="isAutoDetecting" class="detecting-spinner"></span>
      </div>
      
      <!-- API日志 -->
      <div v-if="showApiLogs" class="api-logs">
        <h5>API请求日志:</h5>
        <div v-for="(log, index) in apiLogs" :key="index" class="api-log-entry">
          <span class="log-time">[{{ log.time }}]</span>
          <span :class="['log-type', log.type]">{{ log.type.toUpperCase() }}</span>
          <span class="log-message">{{ log.message }}</span>
        </div>
      </div>
    </div>
    
    <!-- 模拟数据提示 -->
    <div v-if="isUsingMockData && SHOW_API_DIAGNOSTICS" class="mock-data-notice">
      <strong>⚠️ 注意：</strong> 当前显示的是模拟数据，因为无法连接到后端 API。请检查服务器配置或网络连接。
      <div class="error-details">
        <strong>错误信息：</strong> {{ errorMessage }}
      </div>
      <div class="troubleshooting-tips">
        <strong>故障排除建议:</strong>
        <ol>
          <li>检查后端服务是否运行，并查看日志</li>
          <li>检查 <code>internal/server/routes.go</code> 中的 <code>RegisterRoutes</code> 函数，确认API路由正确注册</li>
          <li>注意API路径前缀是否一致，前端使用 <code>/api/xxx</code> 而后端可能只注册了 <code>/xxx</code></li>
          <li>如果使用反向代理，确认代理路径配置正确</li>
        </ol>
      </div>
    </div>

    <!-- 新建任务 -->
    <h4>创建新的任务</h4>
    <div class="form-container">
      <div class="form-group">
        <label>任务名称:</label>
        <input type="text" v-model="newTaskName" placeholder="任务名称" />
      </div>
      
      <div class="form-group">
        <label>数据源:</label>
        <select v-model="newTaskSourceId">
          <option value="">请选择数据源</option>
          <option v-for="source in sources" :key="source.id" :value="source.id">
            {{ source.name }} ({{ source.url }})
          </option>
        </select>
      </div>
      
      <div class="form-group">
        <label>推送模式:</label>
        <div class="push-mode-selection">
          <label>
            <input type="radio" v-model="newTaskPushMode" value="chart" />
            图表模式
          </label>
          <label>
            <input type="radio" v-model="newTaskPushMode" value="text" />
            文本模式
          </label>
        </div>
        <small>图表模式：显示时间序列图表；文本模式：仅显示最新值</small>
      </div>

      <div class="form-group">
        <label>选择预定义 PromQL 查询 (必选):</label>
        <div class="promql-selection">
          <div v-if="promqls.length > 0">
            <!-- 添加全局展开/收起按钮 -->
            <div class="promql-actions">
              <button class="action-btn" @click="toggleAllPromQLs">
                {{ isAllPromQLExpanded ? '收起所有查询' : '展开所有查询' }}
              </button>
            </div>
            <div v-for="promql in promqls" :key="promql.id" class="promql-item">
              <div class="promql-header">
                <input 
                  type="checkbox" 
                  :id="`promql-${promql.id}`" 
                  :value="promql.id.toString()" 
                  v-model="selectedPromQLs"
                >
                <label :for="`promql-${promql.id}`" class="promql-label">
                  {{ promql.name }}
                  <span class="promql-category" v-if="promql.category">({{ promql.category }})</span>
                </label>
                <button class="expand-btn" @click="togglePromQLExpand(promql.id)">
                  {{ expandedPromQLs.includes(promql.id) ? '收起' : '展开' }}
                </button>
              </div>
              <div class="promql-details" :class="{ 'expanded': expandedPromQLs.includes(promql.id) }">
                <pre class="promql-query" v-html="highlightPromQL(promql.query)"></pre>
                <div class="promql-description" v-if="promql.description">
                  {{ promql.description }}
                </div>
                
                <!-- 为选中的 PromQL 添加配置面板 -->
                <div v-if="selectedPromQLs.includes(promql.id.toString())" class="promql-config">
                  <h5>PromQL 配置</h5>
                  <div class="config-group">
                    <label>单位:</label>
                    <input 
                      type="text" 
                      v-model="promqlConfigs[promql.id].unit" 
                      placeholder="例如: MB, GB, %, ms" 
                    />
                  </div>
                  <div class="config-group">
                    <label>指标标签:</label>
                    <select v-model="promqlConfigs[promql.id].metric_label">
                      <option value="pod">Pod 名称</option>
                      <option value="namespace">命名空间</option>
                      <option value="container">容器名称</option>
                      <option value="instance">实例</option>
                      <option value="job">任务名</option>
                      <option value="node">节点名称</option>
                      <option value="cluster">集群名称</option>
                    </select>
                  </div>
                  <div class="config-group">
                    <label>自定义指标标签:</label>
                    <input 
                      type="text" 
                      v-model="promqlConfigs[promql.id].custom_metric_label" 
                      placeholder="留空则使用上方标准标签" 
                    />
                  </div>
                </div>
              </div>
            </div>
          </div>
          <div v-else>
            <p>暂无预定义 PromQL 查询，请先在 PromQL 管理中创建</p>
          </div>
        </div>
        <div class="form-hint">请在 PromQL 管理中创建并选择预定义的查询</div>
      </div>
      
      <div class="form-group">
        <label>时间范围:</label>
        <div class="time-range-inputs">
          <input type="number" v-model="newTaskTimeRangeValue" min="1" />
          <select v-model="newTaskTimeRangeUnit">
            <option value="h">小时</option>
            <option value="d">天</option>
            <option value="M">月</option>
          </select>
        </div>
      </div>
      
      <div class="form-group">
      <label>选择图表模板:
          <select v-model="newTaskChartTemplateId">
            <option value="">-- 请选择图表模板 --</option>
          <option v-for="tmpl in chartTemplates" :key="tmpl.id" :value="tmpl.id">
            {{ tmpl.name }} ({{ tmpl.chart_type }})
          </option>
        </select>
      </label>
    </div>
      <div class="form-group">
      <label>消息标题: <input v-model="newTaskCardTitle"/></label>
    </div>
      <div class="form-group">
    </div>
      <div class="form-group">
      <label>卡片模板:
        <select v-model="newTaskCardTemplate">
          <option value="red">红色</option>
          <option value="carmine">粉色</option>
          <option value="orange">橙色</option>
          <option value="blue">蓝色</option>
          <option value="green">绿色</option>
          <option value="turquoise">青色</option>
          <option value="purple">紫色</option>
          <option value="violet">紫红</option>
          <option value="grey">灰色</option>
        </select>
      </label>
    </div>
      <div class="form-group">
      <label>指标标签:
        <select v-model="newTaskMetricLabel">
          <option value="pod">Pod 名称</option>
          <option value="namespace">命名空间</option>
          <option value="container">容器名称</option>
          <option value="instance">实例</option>
          <option value="job">任务名</option>
          <option value="node">节点名称</option>
          <option value="cluster">集群名称</option>
        </select>
      </label>
    </div>

    <div class="form-group">
      <label>自定义指标标签:
        <input type="text" v-model="newTaskCustomMetricLabel" placeholder="留空则使用标准指标标签" />
      </label>
      <small>如果设置，将覆盖上方选择的标准指标标签</small>
    </div>

      <div class="form-group">
      <label>展示单位:
        <input v-model="newTaskUnit" placeholder="例如: MB, GB, %, ms"/>
      </label>
    </div>

      <div class="form-group">
      <label>按钮文本:
        <input v-model="newTaskButtonText" placeholder="例如: 节点池资源总览"/>
      </label>
      <small>自定义卡片底部按钮的文本，留空则使用默认值</small>
    </div>

      <div class="form-group">
      <label>按钮链接:
        <input v-model="newTaskButtonURL" placeholder="例如: https://grafana.example.com/d/xxx"/>
      </label>
      <small>自定义卡片底部按钮的链接URL，留空则使用默认值</small>
    </div>

      <div class="form-group">
      <label>选择要发送的WebHook(多选):</label>
        <div class="webhook-selection">
          <div v-if="webhooks.length > 0" class="webhook-list">
            <div v-for="webhook in webhooks" :key="webhook.id" class="webhook-item">
              <input 
                type="checkbox" 
                :id="`webhook-${webhook.id}`" 
                :value="webhook.id" 
                v-model="newTaskWebhookIds"
              >
              <label :for="`webhook-${webhook.id}`">{{ webhook.name }}</label>
      </div>
          </div>
          <div v-else class="no-webhooks">
            <p>暂无可用的 WebHook，请先在 WebHook 管理中创建</p>
          </div>
        </div>
        <div class="form-hint">请选择要发送的 WebHook，可多选</div>
    </div>

      <div class="form-group">
        <label>发送时间设置:</label>
        <div class="send-times">
            <div v-for="(time, index) in newTaskSendTimes" :key="index" class="send-time-item">
                <select v-model="time.weekday" class="form-control">
                    <option value="1">周一</option>
                    <option value="2">周二</option>
                    <option value="3">周三</option>
                    <option value="4">周四</option>
                    <option value="5">周五</option>
                    <option value="6">周六</option>
                    <option value="7">周日</option>
                </select>
                <input type="time" v-model="time.send_time" class="form-control" required />
                <button type="button" @click="() => removeSendTime(index, false)" class="remove-time-btn">
                    删除
                </button>
            </div>
            <button type="button" @click="addSendTime" class="add-time-btn">添加发送时间</button>
        </div>
        <div class="form-hint">可以添加多个发送时间，每个时间点都会触发发送</div>
    </div>

      <div class="form-group">
        <label>
          <input type="checkbox" v-model="newTaskShowDataLabel" />
          显示曲线数值
        </label>
        <small>在图表中显示数据点的具体数值</small>
      </div>

      <button @click.prevent="addPushTask" class="create-btn">创建并发送</button>
    </div>
    <hr/>

    <!-- 编辑任务表单 -->
    <div v-if="isEditing" id="edit-task-form" class="form-container edit-form">
      <h4>编辑任务</h4>
      <div class="form-group">
        <label>任务名称:</label>
        <input type="text" v-model="editTaskName" placeholder="任务名称" />
      </div>
      
      <div class="form-group">
        <label>数据源:</label>
        <select v-model="editTaskSourceId">
          <option value="">请选择数据源</option>
          <option v-for="source in sources" :key="source.id" :value="source.id">
            {{ source.name }} ({{ source.url }})
          </option>
        </select>
      </div>
      
      <div class="form-group">
        <label>选择预定义 PromQL 查询 (必选):</label>
        <div class="promql-selection">
          <div v-if="promqls.length > 0">
            <!-- 添加全局展开/收起按钮 -->
            <div class="promql-actions">
              <button class="action-btn" @click="toggleAllPromQLs">
                {{ isAllPromQLExpanded ? '收起所有查询' : '展开所有查询' }}
              </button>
            </div>
            <div v-for="promql in promqls" :key="promql.id" class="promql-item">
              <div class="promql-header">
                <input 
                  type="checkbox" 
                  :id="`edit-promql-${promql.id}`" 
                  :value="promql.id.toString()" 
                  v-model="editSelectedPromQLs"
                >
                <label :for="`edit-promql-${promql.id}`" class="promql-label">
                  {{ promql.name }}
                  <span class="promql-category" v-if="promql.category">({{ promql.category }})</span>
                </label>
                <button class="expand-btn" @click="togglePromQLExpand(promql.id)">
                  {{ expandedPromQLs.includes(promql.id) ? '收起' : '展开' }}
                </button>
              </div>
              <div class="promql-details" :class="{ 'expanded': expandedPromQLs.includes(promql.id) }">
                <pre class="promql-query" v-html="highlightPromQL(promql.query)"></pre>
                <div class="promql-description" v-if="promql.description">
                  {{ promql.description }}
                </div>
              </div>
            </div>
          </div>
          <div v-else>
            <p>暂无预定义 PromQL 查询，请先在 PromQL 管理中创建</p>
          </div>
        </div>
        <div class="form-hint">请在 PromQL 管理中创建并选择预定义的查询</div>
      </div>
      
      <div class="form-group">
        <label>时间范围:</label>
        <div class="time-range-inputs">
          <input type="number" v-model="editTaskTimeRangeValue" min="1" />
          <select v-model="editTaskTimeRangeUnit">
            <option value="h">小时</option>
            <option value="d">天</option>
            <option value="M">月</option>
          </select>
        </div>
      </div>
      
      <div class="form-group">
        <label>选择图表模板:
          <select v-model="editTaskChartTemplateId">
            <option value="">-- 请选择图表模板 --</option>
              <option v-for="tmpl in chartTemplates" :key="tmpl.id" :value="tmpl.id">
                {{ tmpl.name }} ({{ tmpl.chart_type }})
              </option>
            </select>
        </label>
      </div>
      <div class="form-group">
        <label>消息标题: <input v-model="editTaskCardTitle"/></label>
      </div>
      <div class="form-group">
      </div>
      <div class="form-group">
        <label>卡片模板:
            <select v-model="editTaskCardTemplate">
              <option value="red">红色</option>
              <option value="carmine">粉色</option>
              <option value="orange">橙色</option>
              <option value="blue">蓝色</option>
              <option value="green">绿色</option>
              <option value="turquoise">青色</option>
              <option value="purple">紫色</option>
              <option value="violet">紫红</option>
              <option value="grey">灰色</option>
            </select>
        </label>
      </div>
      <div class="form-group">
        <label>指标标签:
            <select v-model="editTaskMetricLabel">
              <option value="pod">Pod 名称</option>
              <option value="namespace">命名空间</option>
              <option value="container">容器名称</option>
              <option value="instance">实例</option>
              <option value="job">任务名</option>
              <option value="node">节点名称</option>
              <option value="cluster">集群名称</option>
            </select>
        </label>
      </div>

      <div class="form-group">
        <label>自定义指标标签:
          <input type="text" v-model="editTaskCustomMetricLabel" placeholder="留空则使用标准指标标签" />
        </label>
        <small>如果设置，将覆盖上方选择的标准指标标签</small>
      </div>

      <div class="form-group">
        <label>展示单位:
            <input v-model="editTaskUnit" placeholder="例如: MB, GB, %, ms"/>
        </label>
      </div>

      <div class="form-group">
        <label>按钮文本:
          <input v-model="editTaskButtonText" placeholder="例如: 节点池资源总览"/>
        </label>
        <small>自定义卡片底部按钮的文本，留空则使用默认值</small>
      </div>

      <div class="form-group">
        <label>按钮链接:
          <input v-model="editTaskButtonURL" placeholder="例如: https://grafana.example.com/d/xxx"/>
        </label>
        <small>自定义卡片底部按钮的链接URL，留空则使用默认值</small>
      </div>

      <div class="form-group">
        <label>选择要发送的WebHook(多选):</label>
        <div class="webhook-selection">
          <div v-if="webhooks.length > 0" class="webhook-list">
            <div v-for="webhook in webhooks" :key="webhook.id" class="webhook-item">
              <input 
                type="checkbox" 
                :id="`edit-webhook-${webhook.id}`" 
                :value="webhook.id" 
                v-model="editTaskWebhookIds"
              >
              <label :for="`edit-webhook-${webhook.id}`">{{ webhook.name }}</label>
            </div>
          </div>
          <div v-else class="no-webhooks">
            <p>暂无可用的 WebHook，请先在 WebHook 管理中创建</p>
          </div>
        </div>
        <div class="form-hint">请选择要发送的 WebHook，可多选</div>
      </div>

      <div class="form-group">
        <label>发送时间设置:</label>
        <div class="send-times">
            <div v-for="(time, index) in editTaskSendTimes" :key="index" class="send-time-item">
                <select v-model="time.weekday" class="form-control">
                    <option value="1">周一</option>
                    <option value="2">周二</option>
                    <option value="3">周三</option>
                    <option value="4">周四</option>
                    <option value="5">周五</option>
                    <option value="6">周六</option>
                    <option value="7">周日</option>
                </select>
                <input type="time" v-model="time.send_time" class="form-control" required />
                <button type="button" @click="() => removeSendTime(index, true)" class="remove-time-btn">
                    删除
                </button>
            </div>
            <button type="button" @click="addSendTime(true)" class="add-time-btn">添加发送时间</button>
        </div>
        <div class="form-hint">可以添加多个发送时间，每个时间点都会触发发送</div>
    </div>

      <div class="form-group">
        <label>
          <input type="checkbox" v-model="editTaskShowDataLabel" />
          显示曲线数值
        </label>
        <small>在图表中显示数据点的具体数值</small>
      </div>

      <div class="edit-actions">
        <button @click.prevent="updateTask" class="update-btn">保存修改</button>
        <button @click.prevent="cancelEdit" class="cancel-btn">取消</button>
      </div>
            </div>

    <!-- 任务列表(可查看已绑定的webhook并编辑/删除) -->
    <div class="task-list">
      <h3>任务列表</h3>
      <div v-if="tasks.length === 0" class="no-tasks">
        <p>暂无任务，请创建新任务</p>
      </div>
      <table v-else>
        <thead>
          <tr>
            <th>ID</th>
            <th>名称</th>
            <th>数据源</th>
            <th>推送模式</th>
            <th>时间范围</th>
            <th>发送时间</th>
            <th>查询/图表标签</th>
            <th>绑定WebHook</th>
            <th>状态</th>
            <th>操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="task in tasks" :key="task.id" :class="{'editing-task': isEditing && editingTaskId === task.id}">
            <td>{{ task.id }}</td>
            <td>{{ task.name }}</td>
            <td>{{ getSourceName(task.source_id) }}</td>
            <td>
              <span :class="['push-mode-badge', task.push_mode || 'chart']">
                {{ task.push_mode === 'text' ? '文本模式' : '图表模式' }}
              </span>
            </td>
            <td>{{ formatTimeRange(task.time_range) }}</td>
            <td>
                <div class="send-times-list">
                    <template v-if="task.send_times && task.send_times.length > 0">
                        <div v-for="(time, index) in task.send_times" :key="index" class="send-time">
                            {{ getWeekdayText(time.weekday) }} {{ time.send_time }}
                        </div>
                    </template>
                    <div v-else class="no-times">
                        未设置发送时间
                    </div>
                </div>
            </td>
            <td>
              <div v-if="task.promql_configs && task.promql_configs.length > 0" class="promql-configs">
                <div v-for="(config, index) in task.promql_configs" :key="index" class="promql-config-item">
                  <span class="promql-name">{{ config.promql_name }}</span>
                  <span v-if="config.unit" class="config-detail">(单位: {{ config.unit }})</span>
                  <span v-if="config.metric_label" class="config-detail">(标签: {{ config.metric_label }})</span>
                </div>
              </div>
              <div v-else-if="task.promql_ids && task.promql_ids.length > 0" class="promql-names">
                <span v-for="(promqlId, index) in task.promql_ids" :key="promqlId" class="promql-tag">
                  {{ getPromqlName(promqlId) }}{{ index < task.promql_ids.length - 1 ? ', ' : '' }}
                </span>
              </div>
              <div v-else class="no-promql-selected">
                <span>未选择</span>
              </div>
            </td>
            <td>
              <div v-if="task.bound_webhooks && task.bound_webhooks.length > 0" class="bound-webhooks">
                <span v-for="(webhook, index) in task.bound_webhooks" :key="webhook.id" class="webhook-tag">
                  {{ webhook.name }}{{ index < task.bound_webhooks.length - 1 ? ', ' : '' }}
                </span>
              </div>
              <div v-else class="no-webhooks-bound">
                <span>未绑定</span>
              </div>
            </td>
            <td>{{ task.enabled ? '启用' : '禁用' }}</td>
            <td>
              <div class="task-actions">
                <button class="edit-btn" @click.prevent="editTask(task)">编辑</button>
                <button class="copy-btn" @click.prevent="copyTask(task)">复制</button>
                <button class="run-btn" @click.prevent="runTask(task.id)">执行</button>
                <button 
                  :class="task.enabled ? 'disable-btn' : 'enable-btn'" 
                  @click.prevent="toggleTask(task.id, !task.enabled)"
                >
                  {{ task.enabled ? '禁用' : '启用' }}
                </button>
                <button class="delete-btn" @click.prevent="deleteTask(task.id)">删除</button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, reactive, watch, onUnmounted } from 'vue'
import { get, post, put, del } from '../utils/api'

// 状态变量
const tasks = ref<any[]>([])
const sources = ref<any[]>([])
const webhooks = ref<any[]>([])
const chartTemplates = ref<any[]>([])
const promqls = ref<any[]>([])
const apiLogs = ref<any[]>([])
const apiTestResults = ref<any[]>([])
const showApiLogs = ref(false)
const apiPathOption = ref('/api')
const customApiPath = ref('')
const recommendedApiPath = ref('')
const isAutoDetecting = ref(false)
const showError = ref(false)
const errorMessage = ref('')
const isUsingMockData = ref(false)
const token = ref('')
const errorTimeout = ref<number | null>(null)
const API_BASE = ref('/api')
const USE_ABSOLUTE_URL = true
const linkWebhookIDs = reactive<{[key: number]: string}>({})

// 配置
const DEBUG = false
const SHOW_API_DIAGNOSTICS = false
const FORCE_MOCK_DATA = false

// 定义debug函数
function debug(message: string, ...args: any[]) {
  if (DEBUG) {
    console.log(`[DEBUG] ${message}`, ...args)
  }
}

// 提供模拟数据
function getMockData() {
  return {
    sources: [
      { id: 1, name: "生产环境", url: "http://prometheus.example.com", type: "prometheus" },
      { id: 2, name: "测试环境", url: "http://test-prometheus.example.com", type: "prometheus" }
    ],
    webhooks: [
      { id: 1, name: "产品通知群", url: "https://open.feishu.cn/webhook/abc123" },
      { id: 2, name: "开发团队群", url: "https://open.feishu.cn/webhook/xyz456" }
    ],
    chartTemplates: [
      { id: 1, name: "折线图", chart_type: "line", template: "{\"type\":\"line\"}" },
      { id: 2, name: "柱状图", chart_type: "bar", template: "{\"type\":\"bar\"}" }
    ],
    promqls: [
      { id: 1, name: "CPU使用率", query: "sum(rate(container_cpu_usage_seconds_total{namespace=\"prod\"}[5m])) by (pod)", category: "资源" },
      { id: 2, name: "内存使用", query: "sum(container_memory_working_set_bytes{namespace=\"prod\"}) by (pod) / (1024*1024)", category: "资源" }
    ],
    tasks: [
      { 
        id: 1, 
        name: "每日CPU报告", 
        source_id: 1, 
        time_range: "24h", 
        schedule_interval: 1440, 
        webhook_ids: [1], 
        bound_webhooks: [{ id: 1, name: "产品通知群" }],
        chart_template_id: 1,
        card_title: "每日CPU使用情况",
        card_template: "blue",
        metric_label: "pod",
        unit: "%",
        enabled: 1,
        send_times: [{ weekday: 1, send_time: "09:00" }],
        promql_ids: [1],
        custom_metric_label: "",
        button_text: "",
        button_url: "",
        show_data_label: false
      }
    ]
  }
}

// 新建任务表单
const newTaskName = ref('')
const newTaskSourceId = ref('')
const newTaskTimeRangeValue = ref('30')
const newTaskTimeRangeUnit = ref('m')
const newTaskChartTemplateId = ref('')
const newTaskCardTitle = ref('')
const newTaskCardTemplate = ref('blue')
const newTaskMetricLabel = ref('pod')
const newTaskCustomMetricLabel = ref('')
const newTaskUnit = ref('')
const newTaskWebhookIds = ref<any[]>([])
const selectedPromQLs = ref([])
const newTaskButtonText = ref('')
const newTaskButtonURL = ref('')
const newTaskShowDataLabel = ref(false) // 添加新的配置项
const newTaskPushMode = ref('chart') // 添加推送模式，默认为图表模式

// 每个 PromQL 的独立配置
const promqlConfigs = ref<Record<number, {
  unit: string
  metric_label: string
  custom_metric_label: string
}>>({})

// 监听 selectedPromQLs 的变化，自动初始化配置
watch(selectedPromQLs, (newVal) => {
  newVal.forEach(id => {
    const numId = parseInt(id)
    if (!promqlConfigs.value[numId]) {
      promqlConfigs.value[numId] = {
        unit: newTaskUnit.value || '',
        metric_label: newTaskMetricLabel.value || 'pod',
        custom_metric_label: ''
      }
    }
  })
})

// 发送时间相关的数据和方法
const taskSendTimes = ref([
    { weekday: 1, send_time: '09:00' }  // 默认周一 09:00
])

// 新增：创建任务的发送时间数组
const newTaskSendTimes = ref([
    { weekday: 1, send_time: '09:00' }  // 默认周一 09:00
])

// 新增：编辑任务的发送时间数组
const editTaskSendTimes = ref([
    { weekday: 1, send_time: '09:00' }  // 默认周一 09:00
])

// 添加发送时间
function addSendTime(isEdit = false) {
    const targetArray = isEdit ? editTaskSendTimes : newTaskSendTimes;
    targetArray.value.push({
        weekday: 1,
        send_time: '09:00'
    })
}

// 删除发送时间
function removeSendTime(index: number, isEdit = false) {
    const targetArray = isEdit ? editTaskSendTimes : newTaskSendTimes;
    targetArray.value = targetArray.value.filter((_, i) => i !== index);
}

// 获取星期几的文字描述
function getWeekdayText(weekday) {
    const weekdays = ['周一', '周二', '周三', '周四', '周五', '周六', '周日']
    return weekdays[weekday - 1] || '未知'
}

// 编辑任务相关状态
const isEditing = ref(false)
const editingTaskId = ref<number | null>(null)
const editTaskName = ref('')
const editTaskSourceId = ref('')
const editTaskTimeRangeValue = ref(1)
const editTaskTimeRangeUnit = ref('h')
const editTaskChartTemplateId = ref('')
const editTaskCardTitle = ref('')
const editTaskCardTemplate = ref('red')
const editTaskMetricLabel = ref('pod')
const editTaskCustomMetricLabel = ref('')
const editTaskUnit = ref('')
const editTaskWebhookIds = ref<any[]>([])
const editSelectedPromQLs = ref<any[]>([])
const editTaskButtonText = ref('')
const editTaskButtonURL = ref('')
const editTaskShowDataLabel = ref(false) // 添加新的配置项

// ============== API路径测试功能 ==============
// 更改API路径
function changeApiPath() {
  API_BASE.value = apiPathOption.value
  debug(`更改API路径前缀为: ${API_BASE.value}`)
}

// 选择自定义路径
function selectCustomPath() {
  apiPathOption.value = 'custom'
  API_BASE.value = customApiPath.value
  debug(`选择自定义路径: ${API_BASE.value}`)
}

// 更新自定义路径
function updateCustomPath() {
  if (apiPathOption.value === 'custom') {
    API_BASE.value = customApiPath.value
    debug(`更新自定义路径为: ${API_BASE.value}`)
  }
}

// 测试API连接
async function testApiConnection() {
  apiTestResults.value = []
  isAutoDetecting.value = true
  debug('测试API连接...')
  
  // 测试端点列表（这些是后端注册的正确端点）
  const endpoints = [
    '/metrics_source',
    '/feishu_webhook',
    '/chart_template',
    '/promqls',
    '/push_task'
  ]
  
  for (const endpoint of endpoints) {
    try {
      const url = getApiUrl(endpoint)
      debug(`测试 ${endpoint} 路径: ${url}`)
      
      const testResult = {
        endpoint,
        url,
        success: false,
        message: '测试中...'
      }
      apiTestResults.value.push(testResult)
      
      // 为push_task端点设置超时
      const timeoutMs = endpoint === '/push_task' ? 15000 : 5000; // push_task给15秒，其他给5秒
      
      // 使用Promise.race实现超时处理
      const fetchPromise = fetch(url);
      const timeoutPromise = new Promise((_, reject) => 
        setTimeout(() => reject(new Error(`请求超时 (${timeoutMs/1000}秒)`)), timeoutMs)
      );
      
      // 记录开始时间用于计算请求耗时
      const startTime = Date.now();
      
      // 发送请求并处理超时
      const response = await Promise.race([fetchPromise, timeoutPromise]);
      const elapsedMs = Date.now() - startTime;
      debug(`${endpoint} 响应状态: ${response.status} ${response.statusText}，耗时: ${elapsedMs}ms`)
      
      // 检查和记录响应头信息（对调试很有用）
      const headers = {};
      response.headers.forEach((value, name) => {
        headers[name] = value;
      });
      debug(`${endpoint} 响应头:`, headers);
      
      // 记录响应头日志（特别是Content-Type很关键）
      apiLogs.value.unshift({
        time: new Date().toLocaleTimeString(),
        type: 'info',
        message: `${endpoint} 响应头: Content-Type=${headers['content-type'] || '未指定'}, Content-Length=${headers['content-length'] || '未知'}`
      });
      
      // 检查响应状态
      if (!response.ok) {
        // 404对于push_task可能是正常的（如果没有任务）
        if (endpoint === '/push_task' && response.status === 404) {
          testResult.success = true
          testResult.message = `返回404 - 可能表示没有任务数据，这是正常的`
          
          apiLogs.value.unshift({
            time: new Date().toLocaleTimeString(),
            type: 'warning',
            message: `${endpoint} 返回404 - 可能是数据库中没有任务数据`
          })
        } else {
          throw new Error(`HTTP ${response.status} ${response.statusText}`);
        }
      } else {
        // 检查响应类型
        const text = await response.text()
        const contentType = response.headers.get('content-type') || '';
        const isHtml = isHtmlResponse(text) || contentType.includes('text/html');
        
        if (isHtml) {
          testResult.success = false
          testResult.message = `返回了HTML而不是JSON，请尝试不同的API路径前缀`
          
          // 添加错误日志
          apiLogs.value.unshift({
            time: new Date().toLocaleTimeString(),
            type: 'error',
            message: `${endpoint} 返回了HTML而不是JSON (Content-Type: ${contentType})`
          })
          
          // 打印HTML内容的前100个字符以便诊断
          const previewHtml = text.substring(0, 100).replace(/\n/g, ' ');
          apiLogs.value.unshift({
            time: new Date().toLocaleTimeString(),
            type: 'info',
            message: `HTML预览: "${previewHtml}${text.length > 100 ? '...' : ''}"`
          })
        } else {
          // 尝试解析JSON
          try {
            const data = JSON.parse(text)
            testResult.success = true
            const itemCount = Array.isArray(data) ? data.length : (data.length !== undefined ? data.length : '非数组');
            testResult.message = `成功！返回了有效的JSON数据 (${itemCount} 项)`
            
            // 检查push_task特殊情况
            if (endpoint === '/push_task') {
              if (Array.isArray(data) && data.length === 0) {
                testResult.message = `成功！返回了空数组 - 没有任务数据`
                
                apiLogs.value.unshift({
                  time: new Date().toLocaleTimeString(),
                  type: 'info',
                  message: `${endpoint} 返回空数组，表示没有任务数据`
                })
                
                apiLogs.value.unshift({
                  time: new Date().toLocaleTimeString(),
                  type: 'info',
                  message: `您可以尝试创建一个新任务来测试完整功能`
                })
              } else {
                // 打印部分数据结构以便调试
                const dataPreview = JSON.stringify(Array.isArray(data) && data.length > 0 ? data[0] : data).substring(0, 100);
                apiLogs.value.unshift({
                  time: new Date().toLocaleTimeString(),
                  type: 'info',
                  message: `${endpoint} 数据结构预览: ${dataPreview}...`
                })
              }
            }
            
            // 添加成功日志
            apiLogs.value.unshift({
              time: new Date().toLocaleTimeString(),
              type: 'success',
              message: `${endpoint} 返回了有效的JSON数据，数据项: ${itemCount}`
            })
          } catch (e) {
            testResult.success = false
            testResult.message = `返回了非HTML内容，但不是有效的JSON: ${e}`
            
            // 添加警告日志
            apiLogs.value.unshift({
              time: new Date().toLocaleTimeString(),
              type: 'warning',
              message: `${endpoint} 返回了非HTML内容，但解析JSON失败: ${e}`
            })
            
            // 打印原始内容前100个字符以便诊断
            const previewText = text.substring(0, 100).replace(/\n/g, ' ');
            apiLogs.value.unshift({
              time: new Date().toLocaleTimeString(),
              type: 'info',
              message: `内容预览: "${previewText}${text.length > 100 ? '...' : ''}"`
            })
          }
        }
      }
    } catch (error) {
      // 更新测试结果
      const testResult = apiTestResults.value.find(r => r.endpoint === endpoint)
      if (testResult) {
        testResult.success = false
        testResult.message = `请求失败: ${error}`
      }
      
      // 添加错误日志
      apiLogs.value.unshift({
        time: new Date().toLocaleTimeString(),
        type: 'error',
        message: `${endpoint} 请求失败: ${error}`
      })
      
      // 对于push_task添加额外提示
      if (endpoint === '/push_task') {
        apiLogs.value.unshift({
          time: new Date().toLocaleTimeString(),
          type: 'info',
          message: `特别注意：push_task端点可能需要额外权限或有特殊处理逻辑`
        })
        
        apiLogs.value.unshift({
          time: new Date().toLocaleTimeString(),
          type: 'info',
          message: `建议检查: 1) 后端日志是否有错误; 2) routes.go中getAllPushTasks函数实现; 3) 数据库是否有push_task表及记录`
        })
      }
    }
  }
  
  // 测试完成
  isAutoDetecting.value = false
  debug('API连接测试完成')
  
  // 如果所有测试都失败，建议使用模拟数据
  const allFailed = apiTestResults.value.every(r => !r.success)
  if (allFailed) {
    apiLogs.value.unshift({
      time: new Date().toLocaleTimeString(),
      type: 'warning',
      message: '所有API端点测试失败，建议使用模拟数据'
    })
  } else {
    const successCount = apiTestResults.value.filter(r => r.success).length
    apiLogs.value.unshift({
      time: new Date().toLocaleTimeString(),
      type: 'info',
      message: `测试完成: ${successCount}/${apiTestResults.value.length} 个端点成功`
    })
  }
  
  // 如果只有push_task失败，显示特殊提示
  const onlyPushTaskFailed = apiTestResults.value.filter(r => !r.success).every(r => r.endpoint === '/push_task');
  const otherEndpointsSucceeded = apiTestResults.value.some(r => r.success);
  
  if (otherEndpointsSucceeded && onlyPushTaskFailed) {
    apiLogs.value.unshift({
      time: new Date().toLocaleTimeString(),
      type: 'warning',
      message: `只有push_task端点失败。这可能是因为数据库中没有任务记录，或者getAllPushTasks函数有特殊的权限要求。`
    })
    
    apiLogs.value.unshift({
      time: new Date().toLocaleTimeString(),
      type: 'info',
      message: `建议: 1) 创建一个push_task看是否解决问题 2) 检查后端日志 3) 检查internal/server/routes.go中的getAllPushTasks函数是否有特殊处理`
    })
  }
}

// 启用模拟数据
function enableMockData() {
  isUsingMockData.value = true
  errorMessage.value = '手动切换到模拟数据模式'
  
  const mockData = getMockData()
  sources.value = mockData.sources
  webhooks.value = mockData.webhooks
  chartTemplates.value = mockData.chartTemplates
  promqls.value = mockData.promqls
  tasks.value = mockData.tasks
  
  debug('已手动切换到模拟数据模式')
  
  // 添加日志
  apiLogs.value.unshift({
    time: new Date().toLocaleTimeString(),
    type: 'info',
    message: '已手动切换到模拟数据模式'
  })
}

// ============== 数据获取函数 ==============
// 获取完整 API URL (自动处理相对/绝对路径)
function getApiUrl(endpoint: string): string {
  // 确保 endpoint 以 / 开头
  if (!endpoint.startsWith('/')) {
    endpoint = '/' + endpoint
  }
  
  // 确保 endpoint 不重复包含 API_BASE
  if (API_BASE.value && endpoint.startsWith(API_BASE.value)) {
    endpoint = endpoint.substring(API_BASE.value.length)
  }

  // 构建完整路径
  const apiPath = `${API_BASE.value}${endpoint}`
  
  if (USE_ABSOLUTE_URL) {
    // 使用绝对路径: http(s)://host:port/api/endpoint
    const protocol = window.location.protocol
    const host = window.location.host
    return `${protocol}//${host}${apiPath}`
  } else {
    // 使用相对路径: /api/endpoint
    return apiPath
  }
}

// 检测响应是否为 HTML (通常表示路由错误)
function isHtmlResponse(text: string): boolean {
  return text.trim().startsWith('<!DOCTYPE html>') || 
         text.trim().startsWith('<html') ||
         text.includes('<head>') || 
         text.includes('<body>')
}

// 从API获取数据，如果失败则使用模拟数据
async function fetchApiData(endpoint: string, mockFallback: any) {
  const url = `${API_BASE.value}${endpoint}`
  debug(`[fetchApiData] 请求: ${url}`)
  
  try {
    apiLogs.value.unshift({
      time: new Date().toLocaleTimeString(),
      type: 'request',
      message: `GET ${url}`
    })
    
    const response = await fetch(url, {
      headers: {
        'Authorization': `Bearer ${token.value}`,
        'Content-Type': 'application/json'
      }
    })
    
    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`)
    }
    
    const data = await response.json()
    
    apiLogs.value.unshift({
      time: new Date().toLocaleTimeString(),
      type: 'success',
      message: `${url} 返回 ${Array.isArray(data) ? data.length : 1} 条数据`
    })
    
    return data
  } catch (error) {
    const fetchError = `获取 ${endpoint} 失败: ${error.message || '未知错误'}`
    debug('[fetchApiData] ' + fetchError)
    errorMessage.value = fetchError
    isUsingMockData.value = true
    
    apiLogs.value.unshift({
      time: new Date().toLocaleTimeString(),
      type: 'error',
      message: fetchError
    })
    
    // 特殊处理push_task
    if (endpoint === '/push_task') {
      apiLogs.value.unshift({
        time: new Date().toLocaleTimeString(),
        type: 'info',
        message: `push_task请求失败提示：您可以尝试手动创建一个任务，或检查后端getAllPushTasks函数是否正常工作`
      })
    }
    
    return mockFallback
  }
}

// 获取所有需要的数据
async function fetchData() {
  debug('开始获取数据...')
  errorMessage.value = ''
  isUsingMockData.value = false
  showError.value = false // 重置错误状态

  try {
    // 获取token
    const storedToken = localStorage.getItem('token')
    if (!storedToken) {
      console.warn('未找到认证信息，将使用模拟数据')
      enableMockData()
      return
    }
    token.value = storedToken

    // 设置请求头
    const headers = {
      'Authorization': `Bearer ${storedToken}`,
      'Content-Type': 'application/json'
    }

    // 并行获取所有数据
    const [sourcesData, webhooksData, templatesData, promqlsData, tasksData] = await Promise.all([
      fetch('/api/metrics_source', { headers }).then(res => {
        if (!res.ok) throw new Error(`获取数据源失败: ${res.status}`)
        return res.json()
      }),
      fetch('/api/feishu_webhook', { headers }).then(res => {
        if (!res.ok) throw new Error(`获取webhook失败: ${res.status}`)
        return res.json()
      }),
      fetch('/api/chart_template', { headers }).then(res => {
        if (!res.ok) throw new Error(`获取图表模板失败: ${res.status}`)
        return res.json()
      }),
      fetch('/api/promqls', { headers }).then(res => {
        if (!res.ok) throw new Error(`获取PromQL失败: ${res.status}`)
        return res.json()
      }),
      fetch('/api/push_task', { headers }).then(res => {
        if (!res.ok) throw new Error(`获取任务列表失败: ${res.status}`)
        return res.json()
      })
    ])

    // 更新数据
    sources.value = Array.isArray(sourcesData) ? sourcesData : []
    webhooks.value = Array.isArray(webhooksData) ? webhooksData : []
    chartTemplates.value = Array.isArray(templatesData) ? templatesData : []
    promqls.value = Array.isArray(promqlsData) ? promqlsData : []

    // 确保任务数据是数组
    if (!Array.isArray(tasksData)) {
      console.error('任务数据格式错误:', tasksData)
      tasks.value = []
      throw new Error('任务数据格式错误')
    }

    // 处理任务数据
    tasks.value = tasksData.map(task => {
      // 确保基本字段存在
      const processedTask = {
        ...task,
        id: task.id || 0,
        name: task.name || '未命名任务',
        source_id: task.source_id || 0,
        enabled: typeof task.enabled === 'boolean' ? task.enabled : task.enabled === 1,
        webhook_ids: Array.isArray(task.webhook_ids) ? task.webhook_ids : [],
        promql_ids: Array.isArray(task.promql_ids) ? task.promql_ids : [],
        send_times: Array.isArray(task.send_times) ? task.send_times : [],
        bound_webhooks: [],
        // 确保 chart_template_id 不会变成 0
        chart_template_id: task.chart_template_id || null
      }

      // 添加调试日志
      console.log('[处理任务数据]', {
        taskId: task.id,
        originalChartTemplateId: task.chart_template_id,
        processedChartTemplateId: processedTask.chart_template_id
      })

      // 处理webhook绑定
      if (Array.isArray(task.webhook_ids)) {
        processedTask.bound_webhooks = task.webhook_ids.map(id => {
          const webhook = webhooks.value.find(w => w.id === id)
          return webhook || { id, name: `WebHook #${id}` }
        })
      }

      // 处理PromQL关联
      if (Array.isArray(task.promql_ids)) {
        processedTask.promql_names = task.promql_ids.map(id => {
          const promql = promqls.value.find(p => p.id === id)
          return promql ? promql.name : `PromQL #${id}`
        })
      }

      return processedTask
    })

    // 记录数据获取结果
    debug('数据获取完成', {
      sourcesCount: sources.value.length,
      webhooksCount: webhooks.value.length,
      templatesCount: chartTemplates.value.length,
      promqlsCount: promqls.value.length,
      tasksCount: tasks.value.length
    })

    // 如果所有数据都是空的，可能是API问题
    if (sources.value.length === 0 && webhooks.value.length === 0 && 
        chartTemplates.value.length === 0 && promqls.value.length === 0 && 
        tasks.value.length === 0) {
      console.warn('所有数据都为空，可能存在API问题')
      throw new Error('无法获取数据，请检查网络连接或API状态')
    }

  } catch (error) {
    console.error('数据获取失败:', error)
    errorMessage.value = error.message || '数据获取失败'
    showError.value = true
    
    // 使用模拟数据作为后备
    if (FORCE_MOCK_DATA) {
      enableMockData()
    } else {
      // 确保所有数据都是空数组而不是 undefined
      sources.value = []
      webhooks.value = []
      chartTemplates.value = []
      promqls.value = []
      tasks.value = []
    }
  }
}

// 取消编辑
function cancelEdit() {
  resetEditForm()
}

// 重置编辑表单
function resetEditForm() {
  isEditing.value = false
  editingTaskId.value = null
  editTaskName.value = ''
  editTaskSourceId.value = ''
  editTaskTimeRangeValue.value = 1
  editTaskTimeRangeUnit.value = 'h'
  editTaskChartTemplateId.value = ''
  editTaskCardTitle.value = ''
  editTaskCardTemplate.value = 'red'
  editTaskMetricLabel.value = 'pod'
  editTaskCustomMetricLabel.value = ''
  editTaskUnit.value = ''
  editTaskWebhookIds.value = []
  editSelectedPromQLs.value = []
  editTaskButtonText.value = ''
  editTaskButtonURL.value = ''
  editTaskShowDataLabel.value = false // 添加新的配置项
  editTaskSendTimes.value = [{ weekday: 1, send_time: '09:00' }]
}

// 切换任务状态
async function toggleTask(taskId: number, enabled: boolean) {
  try {
    const task = tasks.value.find(t => t.id === taskId);
    const taskName = task ? task.name : `任务#${taskId}`;
    const action = enabled ? '启用' : '禁用';
    
    console.log(`[任务${action}] 任务名称:`, taskName)
    console.log(`[任务${action}] 任务ID:`, taskId)
    if (task) {
      console.log(`[任务${action}] 发送时间:`, task.send_times?.map(t => `${getWeekdayText(t.weekday)} ${t.send_time}`).join(', ') || '无')
      console.log(`[任务${action}] 时间范围:`, task.time_range)
    }
    
    if (isUsingMockData.value) {
      await mockApiRequest(`/push_task/${taskId}/toggle`, 'PUT', { enabled });
      
      // 在本地数据中更新任务状态
      if (task) {
        task.enabled = enabled ? 1 : 0;
      }
    } else {
      const url = `${API_BASE.value}/push_task/${taskId}/toggle`;
      const response = await fetch(url, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token.value}`
        },
        body: JSON.stringify({ enabled: enabled ? 1 : 0 })
      });

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }

      const result = await response.json();
      console.log(`[任务${action}] 状态更新成功:`, result)
      
      // 重新获取任务列表
      await fetchData();
    }
    
    console.log(`[任务${action}] 操作成功`)
    
  } catch (error) {
    const errorMsg = `${enabled ? '启用' : '禁用'}任务失败: ${error}`;
    console.error('[任务状态] ' + errorMsg);
    alert(errorMsg);
  }
}

// 立即运行任务
async function runTask(taskId: number) {
  try {
    const task = tasks.value.find(t => t.id === taskId);
    const taskName = task ? task.name : `任务#${taskId}`;
    
    console.log('[任务执行] 任务名称:', taskName);
    console.log('[任务执行] 任务ID:', taskId);
    if (task) {
      console.log('[任务执行] 发送时间:', task.send_times?.map(t => `${getWeekdayText(t.weekday)} ${t.send_time}`).join(', ') || '无');
      console.log('[任务执行] 时间范围:', task.time_range);
    }
    
    if (isUsingMockData.value) {
      await mockApiRequest(`/push_task/${taskId}/run`, 'POST');
      console.log('[任务执行] 已触发执行 (模拟模式)');
      alert('任务已触发运行 (模拟模式)');
    } else {
      const url = `${API_BASE.value}/push_task/${taskId}/run`;
      const response = await fetch(url, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token.value}`
        }
      });

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }

      const result = await response.json();
      console.log('[任务执行] 执行结果:', result);
      console.log('[任务执行] 已触发执行');
      alert('任务已触发运行');
    }
  } catch (error) {
    const errorMsg = `运行任务失败: ${error}`;
    console.error('[任务执行] ' + errorMsg);
    alert(errorMsg);
  }
}

// 删除任务
async function deleteTask(taskId: number) {
  if (!confirm('确定要删除此任务吗？')) {
    return
  }

  try {
    debug('删除任务:', taskId)
    
    if (isUsingMockData.value) {
      await mockApiRequest(`/push_task/${taskId}`, 'DELETE');
      
      // 在本地数据中删除任务
      tasks.value = tasks.value.filter(t => t.id !== taskId);
    } else {
      const url = `${API_BASE.value}/push_task/${taskId}`;
      await del(url);
      fetchData();
    }
  } catch (error) {
    debug('删除任务失败:', error)
    alert('删除任务失败')
  }
}

// 自动检测API路径
async function autoDetectApiPath() {
  if (isAutoDetecting.value) return;
  isAutoDetecting.value = true;
  recommendedApiPath.value = '';
  apiTestResults.value = [];
  
  debug('开始自动检测API路径前缀...');
  
  // 测试不同的前缀
  const prefixes = ['', '/api', '/api/v1', '/v1', '/api/v2', '/v2'];
  // 正确的后端端点（与routes.go匹配）
  const endpoints = [
    '/metrics_source',
    '/feishu_webhook', 
    '/chart_template',
    '/promqls',
    '/push_task'
  ];
  
  let bestPrefix = '';
  let bestSuccessCount = 0;
  
  // 测试每个前缀
  for (const prefix of prefixes) {
    debug(`测试前缀: "${prefix}"`);
    let successCount = 0;
    
    // 临时设置API前缀
    const originalPrefix = API_BASE.value;
    API_BASE.value = prefix;
    
    // 测试每个端点
    for (const endpoint of endpoints) {
      try {
        const url = getApiUrl(endpoint);
        debug(`测试URL: ${url}`);
        
        // 发送请求
        const response = await fetch(url);
        const text = await response.text();
        
        // 检查是否为HTML
        if (!isHtmlResponse(text)) {
          try {
            // 尝试解析JSON
            JSON.parse(text);
            successCount++;
            debug(`${endpoint} 在前缀 "${prefix}" 下成功`);
          } catch (e) {
            debug(`${endpoint} 返回非HTML但不是有效JSON: ${e}`);
          }
        }
      } catch (error) {
        debug(`测试 ${endpoint} 失败: ${error}`);
      }
    }
    
    debug(`前缀 "${prefix}" 成功数: ${successCount}/${endpoints.length}`);
    
    // 如果这个前缀的成功率更高，更新推荐前缀
    if (successCount > bestSuccessCount) {
      bestSuccessCount = successCount;
      bestPrefix = prefix;
    }
  }
  
  // 恢复原来的API前缀
  API_BASE.value = apiPathOption.value;
  
  // 设置推荐的API路径
  if (bestSuccessCount > 0) {
    recommendedApiPath.value = bestPrefix;
    debug(`推荐API前缀: "${bestPrefix}" (${bestSuccessCount}/${endpoints.length} 成功)`);
  } else {
    debug('未找到有效的API前缀');
  }
  
  isAutoDetecting.value = false;
}

// 应用推荐的API路径
function applyRecommendedPath() {
  if (recommendedApiPath.value) {
    API_BASE.value = recommendedApiPath.value;
    apiPathOption.value = recommendedApiPath.value;
    debug(`已应用推荐的API前缀: "${recommendedApiPath.value}"`);
    
    // 重新测试
    testApiConnection();
  }
}

// 获取周几名称
function getWeekdayName(weekday: number): string {
  const weekdays = ['周一', '周二', '周三', '周四', '周五', '周六', '周日']
  return weekdays[(weekday - 1) % 7] || '未知'
}

// 更新任务
async function updateTask() {
  try {
    if (!editTaskName.value) {
      alert('请输入任务名称')
      return
    }
    if (!editTaskSourceId.value) {
      alert('请选择数据源')
      return
    }
    if (editSelectedPromQLs.value.length === 0) {
      alert('请至少选择一个PromQL查询')
      return
    }
    if (editTaskSendTimes.value.length === 0) {
      alert('请至少设置一个发送时间')
      return
    }
    
    const oldTask = tasks.value.find(t => t.id === editingTaskId.value)
    
    // 添加任务更新前的日志
    console.log('[任务更新] 开始更新任务...')
    console.log('[任务更新] 任务ID:', editingTaskId.value)
    console.log('[任务更新] 任务名称:', editTaskName.value)
    console.log('[任务更新] 原发送时间:', oldTask?.send_times?.map(t => `${getWeekdayText(t.weekday)} ${t.send_time}`).join(', ') || '无')
    console.log('[任务更新] 原时间范围:', oldTask?.time_range || '无')
    
    const payload = {
      id: editingTaskId.value,
      name: editTaskName.value,
      source_id: parseInt(editTaskSourceId.value),
      promql_ids: editSelectedPromQLs.value.map(id => parseInt(id)),
      query: editSelectedPromQLs.value.map(id => {
        const promql = promqls.value.find(p => p.id.toString() === id)
        return promql ? promql.query : ''
      }).filter(q => q).join(', '),
      time_range: `${editTaskTimeRangeValue.value}${editTaskTimeRangeUnit.value}`,
      step: calculateStep(editTaskTimeRangeValue.value, editTaskTimeRangeUnit.value),
      chart_template_id: parseInt(editTaskChartTemplateId.value),
      webhook_ids: editTaskWebhookIds.value.map(id => parseInt(id)),
      card_title: editTaskCardTitle.value,
      card_template: editTaskCardTemplate.value,
      metric_label: editTaskMetricLabel.value,
      custom_metric_label: editTaskCustomMetricLabel.value,
      unit: editTaskUnit.value,
      button_text: editTaskButtonText.value,
      button_url: editTaskButtonURL.value,
      show_data_label: editTaskShowDataLabel.value,
      enabled: 1,
      send_times: editTaskSendTimes.value.map(time => ({
        weekday: parseInt(time.weekday),
        send_time: time.send_time
      }))
    }

    if (isUsingMockData.value) {
      // 在本地更新数据
      const taskIndex = tasks.value.findIndex(t => t.id === editingTaskId.value)
      if (taskIndex !== -1) {
        const updatedTask = {
          ...tasks.value[taskIndex],
          ...payload,
          bound_webhooks: payload.webhook_ids.map(id => {
            const webhook = webhooks.value.find(w => w.id === id)
            return webhook || { id, name: `WebHook #${id}` }
          })
        }
        tasks.value.splice(taskIndex, 1, updatedTask)
      }
      
      console.log('[任务更新] 更新成功 (模拟模式)')
      resetEditForm()
      alert('任务更新成功')
    } else {
      const url = `${API_BASE.value}/push_task/${editingTaskId.value}`
      console.log('[任务更新] 发送请求到:', url)
      
      const response = await fetch(url, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token.value}`
        },
        body: JSON.stringify(payload)
      })

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`)
      }

      const result = await response.json()
      console.log('[任务更新] 服务器响应:', result)
      
      // 重新获取任务列表
      await fetchData()
      // 重置编辑状态
      resetEditForm()
      alert('任务更新成功')
    }
  } catch (error) {
    const errorMsg = `更新任务失败: ${error}`
    console.error('[任务更新] ' + errorMsg)
    alert(errorMsg)
  }
}

// 生命周期钩子
onMounted(async () => {
  debug('组件已挂载，开始初始化...')
  
  // 设置定期刷新的间隔
  let refreshInterval: number | null = null

  try {
    // 获取数据
    await fetchData()
    
    // 设置定期刷新
    refreshInterval = window.setInterval(() => {
      fetchData()
    }, 30000) // 每30秒刷新一次

    // 初始化 MutationObserver
    const targetNode = document.querySelector('.tab-content')
    if (targetNode) {
      const observer = new MutationObserver((mutations) => {
        mutations.forEach((mutation) => {
          if (mutation.type === 'childList') {
            // 处理DOM变化
            debug('DOM变化:', mutation)
          }
        })
      })

      observer.observe(targetNode, {
        childList: true,
        subtree: true
      })

      // 在组件卸载时断开观察器
      onUnmounted(() => {
        observer.disconnect()
        if (refreshInterval) {
          clearInterval(refreshInterval)
        }
        if (errorTimeout.value) {
          clearTimeout(errorTimeout.value)
        }
      })
    } else {
      console.warn('未找到目标节点，MutationObserver未初始化')
    }
  } catch (error) {
    console.error('初始化失败:', error)
    errorMessage.value = '初始化失败，请刷新页面重试'
    showError.value = true
  }
})

// 计算步长的辅助函数
function calculateStep(value: number | string, unit: string): number {
  const numValue = Number(value)
  if (isNaN(numValue)) return 300 // 默认5分钟

  switch (unit) {
    case 'h':
      return numValue * 3600
    case 'd':
      return numValue * 86400
    case 'M':
      return numValue * 2592000 // 30天
    default:
      return numValue * 60 // 默认按分钟计算
  }
}

// 格式化时间范围的辅助函数
function formatTimeRange(timeRange: string): string {
  if (!timeRange) return '未设置'
  
  const match = timeRange.match(/^(\d+)([hHdDmM])$/)
  if (!match) return timeRange
  
  const [_, value, unit] = match
  const unitMap = {
    h: '小时',
    H: '小时',
    d: '天',
    D: '天',
    m: '分钟',
    M: '月'
  }
  
  return `${value}${unitMap[unit] || unit}`
}

// 获取数据源名称的辅助函数
function getSourceName(sourceId: number): string {
  const source = sources.value.find(s => s.id === sourceId)
  return source ? source.name : `数据源#${sourceId}`
}

// 获取PromQL名称的辅助函数
function getPromqlName(promqlId: number): string {
  const promql = promqls.value.find(p => p.id === promqlId)
  return promql ? promql.name : `PromQL#${promqlId}`
}

// 模拟API请求的辅助函数
async function mockApiRequest(endpoint: string, method: string, data?: any) {
  console.log(`[Mock API] ${method} ${endpoint}`, data)
  await new Promise(resolve => setTimeout(resolve, 500)) // 模拟网络延迟
  return { success: true }
}

// 编辑任务
function editTask(task) {
  isEditing.value = true
  editingTaskId.value = task.id
  editTaskName.value = task.name
  editTaskSourceId.value = task.source_id.toString()
  
  // 解析时间范围
  const timeRangeMatch = task.time_range.match(/^(\d+)([hHdDmM])$/)
  if (timeRangeMatch) {
    editTaskTimeRangeValue.value = parseInt(timeRangeMatch[1])
    editTaskTimeRangeUnit.value = timeRangeMatch[2].toLowerCase()
  }
  
  // 确保图表模板ID被正确设置
  if (task.chart_template_id) {
    editTaskChartTemplateId.value = task.chart_template_id.toString()
    console.log('[编辑任务] 设置图表模板ID:', editTaskChartTemplateId.value, '原始值:', task.chart_template_id)
  } else {
    editTaskChartTemplateId.value = ''
    console.log('[编辑任务] 任务没有图表模板ID')
  }
  
  editTaskCardTitle.value = task.card_title || ''
  editTaskCardTemplate.value = task.card_template || 'blue'
  editTaskMetricLabel.value = task.metric_label || 'pod'
  editTaskCustomMetricLabel.value = task.custom_metric_label || ''
  editTaskUnit.value = task.unit || ''
  editTaskWebhookIds.value = task.webhook_ids || []
  editSelectedPromQLs.value = task.promql_ids?.map(id => id.toString()) || []
  editTaskButtonText.value = task.button_text || ''
  editTaskButtonURL.value = task.button_url || ''
  editTaskShowDataLabel.value = task.show_data_label || false
  
  // 设置发送时间
  if (Array.isArray(task.send_times) && task.send_times.length > 0) {
    editTaskSendTimes.value = task.send_times.map(time => ({
      weekday: time.weekday,
      send_time: time.send_time
    }))
  } else {
    editTaskSendTimes.value = [{ weekday: 1, send_time: '09:00' }]
  }
  
  // 添加调试日志
  console.log('[编辑任务] 任务数据:', {
    id: task.id,
    name: task.name,
    chart_template_id: task.chart_template_id,
    promql_ids: task.promql_ids,
    webhook_ids: task.webhook_ids,
    send_times: task.send_times,
    editTaskChartTemplateId: editTaskChartTemplateId.value
  })
  
  // 滚动到编辑表单
  setTimeout(() => {
    const editForm = document.getElementById('edit-task-form')
    if (editForm) {
      editForm.scrollIntoView({ behavior: 'smooth' })
    }
  }, 100)
}

// 创建新任务
async function addPushTask() {
  try {
    // 验证必填字段
    if (!newTaskName.value) {
      alert('请输入任务名称')
      return
    }
    if (!newTaskSourceId.value) {
      alert('请选择数据源')
      return
    }
    if (selectedPromQLs.value.length === 0) {
      alert('请至少选择一个PromQL查询')
      return
    }
    if (newTaskSendTimes.value.length === 0) {
      alert('请至少设置一个发送时间')
      return
    }
    if (!newTaskChartTemplateId.value) {
      alert('请选择图表模板')
      return
    }

    // 构建 PromQL 配置列表
    const promqlConfigsList = selectedPromQLs.value.map(id => {
      const numId = parseInt(id)
      const config = promqlConfigs.value[numId] || {
        unit: '',
        metric_label: 'pod',
        custom_metric_label: ''
      }
      return {
        promql_id: numId,
        unit: config.unit,
        metric_label: config.metric_label,
        custom_metric_label: config.custom_metric_label,
        chart_template_id: parseInt(newTaskChartTemplateId.value) || null
      }
    })

    // 构建请求数据
    const payload = {
      name: newTaskName.value,
      source_id: parseInt(newTaskSourceId.value),
      promql_ids: selectedPromQLs.value.map(id => parseInt(id)),
      promql_configs: promqlConfigsList, // 新增：每个 PromQL 的独立配置
      query: selectedPromQLs.value.map(id => {
        const promql = promqls.value.find(p => p.id.toString() === id)
        return promql ? promql.query : ''
      }).filter(q => q).join(', '),
      time_range: `${newTaskTimeRangeValue.value}${newTaskTimeRangeUnit.value}`,
      step: calculateStep(newTaskTimeRangeValue.value, newTaskTimeRangeUnit.value),
      chart_template_id: parseInt(newTaskChartTemplateId.value) || null, // 确保不会变成0
      webhook_ids: newTaskWebhookIds.value.map(id => parseInt(id)),
      card_title: newTaskCardTitle.value,
      card_template: newTaskCardTemplate.value,
      metric_label: newTaskMetricLabel.value,
      custom_metric_label: newTaskCustomMetricLabel.value,
      unit: newTaskUnit.value,
      button_text: newTaskButtonText.value,
      button_url: newTaskButtonURL.value,
      show_data_label: newTaskShowDataLabel.value,
      push_mode: newTaskPushMode.value, // 新增：推送模式
      enabled: 1,
      send_times: newTaskSendTimes.value.map(time => ({
        weekday: parseInt(time.weekday),
        send_time: time.send_time
      })),
      // 为每个 PromQL 查询添加图表模板关联
      promql_chart_templates: selectedPromQLs.value.map(promqlId => ({
        promql_id: parseInt(promqlId),
        chart_template_id: parseInt(newTaskChartTemplateId.value) || null // 确保不会变成0
      }))
    }

    console.log('[创建任务] 开始创建任务...')
    console.log('[创建任务] 任务名称:', newTaskName.value)
    console.log('[创建任务] 发送时间:', newTaskSendTimes.value.map(t => `${getWeekdayText(t.weekday)} ${t.send_time}`).join(', '))
    console.log('[创建任务] 时间范围:', `${newTaskTimeRangeValue.value}${newTaskTimeRangeUnit.value}`)
    console.log('[创建任务] 图表模板ID:', newTaskChartTemplateId.value)
    console.log('[创建任务] PromQL IDs:', selectedPromQLs.value)
    console.log('[创建任务] 图表模板关联:', payload.promql_chart_templates)
    console.log('[创建任务] 完整请求数据:', payload)

    if (isUsingMockData.value) {
      // 模拟创建任务
      await mockApiRequest('/push_task', 'POST', payload)
      
      // 生成模拟ID
      const newId = tasks.value.length > 0 ? Math.max(...tasks.value.map(t => t.id)) + 1 : 1
      
      // 在本地添加任务
      const newTask = {
        ...payload,
        id: newId,
        bound_webhooks: payload.webhook_ids.map(id => {
          const webhook = webhooks.value.find(w => w.id === id)
          return webhook || { id, name: `WebHook #${id}` }
        })
      }
      tasks.value.push(newTask)
      
      console.log('[创建任务] 创建成功 (模拟模式)')
      alert('任务创建成功')
    } else {
      const url = `${API_BASE.value}/push_task`
      console.log('[创建任务] 发送请求到:', url)
      console.log('[创建任务] 请求数据:', payload)
      
      const response = await fetch(url, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token.value}`
        },
        body: JSON.stringify(payload)
      })

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`)
      }

      const result = await response.json()
      console.log('[创建任务] 服务器响应:', result)
      
      // 重新获取任务列表
      await fetchData()
      alert('任务创建成功')
    }

    // 重置表单
    newTaskName.value = ''
    newTaskSourceId.value = ''
    newTaskTimeRangeValue.value = '30'
    newTaskTimeRangeUnit.value = 'm'
    newTaskChartTemplateId.value = ''
    newTaskCardTitle.value = ''
    newTaskCardTemplate.value = 'blue'
    newTaskMetricLabel.value = 'pod'
    newTaskCustomMetricLabel.value = ''
    newTaskUnit.value = ''
    newTaskPushMode.value = 'chart' // 重置推送模式
    promqlConfigs.value = {} // 重置 PromQL 配置
    newTaskWebhookIds.value = []
    selectedPromQLs.value = []
    newTaskButtonText.value = ''
    newTaskButtonURL.value = ''
    newTaskShowDataLabel.value = false
    newTaskSendTimes.value = [{ weekday: 1, send_time: '09:00' }]

  } catch (error) {
    const errorMsg = `创建任务失败: ${error}`
    console.error('[创建任务] ' + errorMsg)
    alert(errorMsg)
  }
}

// 添加复制任务功能
async function copyTask(task) {
  try {
    // 验证原任务的图表模板ID
    if (!task.chart_template_id) {
      alert('原任务没有设置图表模板，请先设置图表模板')
      return
    }

    // 验证图表模板是否存在
    const templateExists = chartTemplates.value.some(t => t.id === parseInt(task.chart_template_id))
    if (!templateExists) {
      alert('原任务的图表模板不存在，请先检查图表模板配置')
      return
    }

    // 获取所有相关的 PromQL 查询
    let queries = [];
    let promqlChartTemplates = [];
    if (task.promql_ids && task.promql_ids.length > 0) {
      // 获取 PromQL 查询
      queries = promqls.value
        .filter(p => task.promql_ids.includes(p.id))
        .map(p => p.query);
      
      // 为每个 PromQL 查询创建图表模板关联
      promqlChartTemplates = task.promql_ids.map(promqlId => ({
        promql_id: parseInt(promqlId),
        chart_template_id: parseInt(task.chart_template_id)
      }));

      console.log('[复制任务] PromQL查询数量:', queries.length)
      console.log('[复制任务] 图表模板关联数量:', promqlChartTemplates.length)
    } else {
      alert('原任务没有关联的PromQL查询，请先配置PromQL查询')
      return
    }

    // 创建新任务对象，复制原任务的所有属性
    const newTask = {
      name: `${task.name} (复制)`,
      source_id: parseInt(task.source_id),
      promql_ids: task.promql_ids.map(id => parseInt(id)),
      query: queries.join(', '), // 设置 query 字段为所有 PromQL 查询的组合
      time_range: task.time_range,
      step: task.step,
      chart_template_id: parseInt(task.chart_template_id),
      webhook_ids: task.webhook_ids.map(id => parseInt(id)),
      card_title: task.card_title || '',
      card_template: task.card_template || 'blue',
      metric_label: task.metric_label || 'pod',
      custom_metric_label: task.custom_metric_label || '',
      unit: task.unit || '',
      button_text: task.button_text || '',
      button_url: task.button_url || '',
      show_data_label: task.show_data_label || false,
      enabled: 1,
      send_times: task.send_times.map(time => ({
        weekday: parseInt(time.weekday),
        send_time: time.send_time
      })),
      schedule_interval: task.schedule_interval || 3600,
      promql_chart_templates: promqlChartTemplates // 使用新创建的图表模板关联
    }

    console.log('[复制任务] 开始复制任务...')
    console.log('[复制任务] 原任务名称:', task.name)
    console.log('[复制任务] 新任务名称:', newTask.name)
    console.log('[复制任务] 原任务图表模板ID:', task.chart_template_id)
    console.log('[复制任务] 图表模板关联:', promqlChartTemplates)
    console.log('[复制任务] PromQL IDs:', newTask.promql_ids)
    console.log('[复制任务] 请求数据:', newTask)

    if (isUsingMockData.value) {
      // 模拟创建任务
      await mockApiRequest('/push_task', 'POST', newTask)
      
      // 生成模拟ID
      const newId = tasks.value.length > 0 ? Math.max(...tasks.value.map(t => t.id)) + 1 : 1
      
      // 在本地添加任务
      const createdTask = {
        ...newTask,
        id: newId,
        bound_webhooks: newTask.webhook_ids.map(id => {
          const webhook = webhooks.value.find(w => w.id === id)
          return webhook || { id, name: `WebHook #${id}` }
        })
      }
      tasks.value.push(createdTask)
      
      console.log('[复制任务] 复制成功 (模拟模式)')
      alert('任务复制成功')
    } else {
      const url = `${API_BASE.value}/push_task`
      console.log('[复制任务] 发送请求到:', url)
      
      const response = await fetch(url, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token.value}`
        },
        body: JSON.stringify(newTask)
      })

      if (!response.ok) {
        const errorText = await response.text()
        throw new Error(`HTTP error! status: ${response.status}, message: ${errorText}`)
      }

      const result = await response.json()
      console.log('[复制任务] 服务器响应:', result)
      
      // 重新获取任务列表
      await fetchData()
      alert('任务复制成功')
    }
  } catch (error) {
    const errorMsg = `复制任务失败: ${error}`
    console.error('[复制任务] ' + errorMsg)
    alert(errorMsg)
  }
}

// 添加 PromQL 展开状态管理
const expandedPromQLs = ref<number[]>([])
const isAllPromQLExpanded = ref(false)

// 展开/收起所有 PromQL 查询
function toggleAllPromQLs() {
  if (isAllPromQLExpanded.value) {
    expandedPromQLs.value = []
  } else {
    expandedPromQLs.value = promqls.value.map(p => p.id)
  }
  isAllPromQLExpanded.value = !isAllPromQLExpanded.value
}

// 展开/收起单个 PromQL 查询
function togglePromQLExpand(id: number) {
  const index = expandedPromQLs.value.indexOf(id)
  if (index === -1) {
    expandedPromQLs.value.push(id)
  } else {
    expandedPromQLs.value.splice(index, 1)
  }
  // 更新全局展开状态
  isAllPromQLExpanded.value = expandedPromQLs.value.length === promqls.value.length
}

// PromQL 语法高亮函数
function highlightPromQL(query: string): string {
  if (!query) return ''
  
  // 基本的 PromQL 关键字和函数
  const keywords = [
    'sum', 'rate', 'irate', 'avg', 'max', 'min', 'count',
    'by', 'without', 'offset', 'bool', 'and', 'or', 'unless',
    'group', 'ignoring', 'on', 'topk', 'bottomk'
  ]
  
  // 转义 HTML 特殊字符
  let highlighted = query.replace(/[&<>]/g, char => {
    const entities: { [key: string]: string } = {
      '&': '&amp;',
      '<': '&lt;',
      '>': '&gt;'
    }
    return entities[char] || char
  })
  
  // 高亮关键字
  keywords.forEach(keyword => {
    const regex = new RegExp(`\\b${keyword}\\b`, 'g')
    highlighted = highlighted.replace(regex, `<span class="keyword">${keyword}</span>`)
  })
  
  // 高亮标签和值
  highlighted = highlighted.replace(
    /(\{[^}]*\})/g,
    (match) => `<span class="label">${match}</span>`
  )
  
  // 高亮数字
  highlighted = highlighted.replace(
    /\b(\d+(\.\d+)?)\b/g,
    '<span class="number">$1</span>'
  )
  
  // 高亮时间单位
  highlighted = highlighted.replace(
    /\b(\d+)(s|m|h|d|w|y)\b/g,
    '<span class="number">$1</span><span class="unit">$2</span>'
  )
  
  return highlighted
}
</script>

<style scoped>
.tab-content {
  border:1px solid #ccc;
  padding:16px;
  margin-bottom:24px;
}
table {
  margin-top:10px;
  width:100%;
  border-collapse:collapse;
}
th,td {
  border:1px solid #ccc;
  padding:4px 8px;
  text-align:left;
}

.webhook-item {
  display: flex;
  align-items: center;
  margin-bottom: 4px;
}

.webhook-item .small-btn {
  margin-left: 8px;
  padding: 2px 4px;
  font-size: 12px;
}

.add-webhook {
  margin-top: 8px;
  padding-top: 8px;
  border-top: 1px dashed #ccc;
}

.no-webhooks {
  color: #999;
  font-style: italic;
  margin-bottom: 8px;
}

.webhook-selection {
  max-height: 200px;
  overflow-y: auto;
  border: 1px solid #ddd;
  border-radius: 4px;
  padding: 8px;
  margin-bottom: 8px;
}

.webhook-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.webhook-item {
  display: flex;
  align-items: center;
  gap: 8px;
}

.no-webhooks, .no-tasks {
  padding: 12px;
  background-color: #f8f9fa;
  border-radius: 4px;
  text-align: center;
  color: #6c757d;
}

.bound-webhooks {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
}

.webhook-tag {
  font-size: 0.9em;
  color: #0056b3;
}

.no-webhooks-bound {
  color: #6c757d;
  font-style: italic;
}

.promql-tag {
  font-size: 0.9em;
  color: #28a745;
  background-color: #f0f9f0;
  padding: 2px 6px;
  border-radius: 3px;
  margin-right: 3px;
  display: inline-block;
  margin-bottom: 3px;
}

.promql-names {
  display: flex;
  flex-wrap: wrap;
  max-width: 200px;
  gap: 3px;
}

.no-promql-selected {
  color: #6c757d;
  font-style: italic;
}

.task-actions {
  display: flex;
  gap: 4px;
  flex-wrap: wrap;
}

.task-actions button {
  padding: 4px 8px;
  font-size: 0.9em;
}

.edit-btn {
  background-color: #17a2b8;
  color: white;
  border-color: #17a2b8;
}

.run-btn {
  background-color: #28a745;
  color: white;
  border-color: #28a745;
}

.disable-btn {
  background-color: #ffc107;
  color: #212529;
  border-color: #ffc107;
}

.enable-btn {
  background-color: #6c757d;
  color: white;
  border-color: #6c757d;
}

.delete-btn {
  background-color: #dc3545;
  color: white;
  border-color: #dc3545;
}

.copy-btn {
  background-color: #6610f2;
  color: white;
  border-color: #6610f2;
  padding: 4px 8px;
  font-size: 0.9em;
  border-radius: 4px;
  cursor: pointer;
  transition: background-color 0.2s;
}

.copy-btn:hover {
  background-color: #520dc2;
  border-color: #520dc2;
}

.task-actions button:hover {
  opacity: 0.9;
}

.time-range-inputs {
  display: flex;
  gap: 10px;
}

.time-range-inputs input {
  width: 80px;
}

.create-btn {
  margin-top: 15px;
  padding: 10px 20px;
  background-color: #4CAF50;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 16px;
  font-weight: bold;
}

.create-btn:hover {
  background-color: #45a049;
}

input:required, select:required, textarea:required {
  border-left: 3px solid #4CAF50;
}

input:invalid, select:invalid, textarea:invalid {
  border-left: 3px solid #f44336;
}

.form-group {
  margin-bottom: 15px;
}

.help-text {
  margin-top: 5px;
  font-size: 12px;
  color: #666;
  display: flex;
  align-items: flex-start;
}

.info-icon {
  margin-right: 5px;
  color: #007bff;
}

.required {
  color: red;
  margin-left: 3px;
}

.promql-selection {
  max-height: 400px;
  overflow-y: auto;
  border: 1px solid #ddd;
  border-radius: 4px;
  padding: 15px;
  background: #fff;
}

.promql-actions {
  margin-bottom: 15px;
  display: flex;
  justify-content: flex-end;
}

.promql-item {
  margin-bottom: 12px;
  padding: 8px;
  border: 1px solid #eee;
  border-radius: 4px;
  background: #f8f9fa;
}

.promql-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
}

.promql-label {
  flex: 1;
  font-weight: 500;
  cursor: pointer;
}

.promql-category {
  color: #666;
  font-size: 0.9em;
  font-weight: normal;
  margin-left: 8px;
}

.promql-details {
  display: none;
  margin-top: 8px;
  padding: 8px;
  background: #fff;
  border-radius: 4px;
}

.promql-details.expanded {
  display: block;
}

.promql-query {
  margin: 0;
  padding: 8px;
  background: #f8f9fa;
  border-radius: 4px;
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', 'Consolas', monospace;
  font-size: 13px;
  line-height: 1.5;
  white-space: pre-wrap;
  word-break: break-word;
}

.promql-description {
  margin-top: 8px;
  color: #666;
  font-size: 0.9em;
  line-height: 1.4;
}

.expand-btn {
  padding: 4px 8px;
  background: #e9ecef;
  border: none;
  border-radius: 3px;
  font-size: 12px;
  cursor: pointer;
  opacity: 0.8;
  transition: all 0.2s ease;
}

.expand-btn:hover {
  opacity: 1;
  background: #dee2e6;
}

.action-btn {
  padding: 6px 12px;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 14px;
  transition: all 0.3s ease;
  background-color: #f0f0f0;
}

.action-btn:hover {
  background-color: #e0e0e0;
}

/* PromQL 语法高亮样式 */
:deep(.keyword) {
  color: #0066cc;
  font-weight: 500;
}

:deep(.label) {
  color: #e83e8c;
}

:deep(.number) {
  color: #2e7d32;
}

:deep(.unit) {
  color: #0066cc;
  font-weight: 500;
}

.form-hint {
  font-size: 0.85em;
  color: #666;
  margin-top: 5px;
}

.task-promql {
  margin-bottom: 3px;
  font-size: 0.9em;
  color: #007bff;
}

.no-data {
  padding: 20px;
  text-align: center;
  color: #666;
  font-style: italic;
  border: 1px dashed #ccc;
  margin: 20px 0;
}

/* 模拟数据提示样式 */
.mock-data-notice {
  background-color: #fff3cd;
  color: #856404;
  padding: 10px 15px;
  margin-bottom: 20px;
  border: 1px solid #ffeeba;
  border-radius: 4px;
  position: relative;
}

.mock-data-notice strong {
  font-weight: bold;
}

.error-details {
  margin-top: 8px;
  font-size: 0.9em;
  border-top: 1px dashed #ffeeba;
  padding-top: 8px;
}

/* 模拟数据警告样式 */
.mock-data-alert {
  display: flex;
  margin-bottom: 20px;
  padding: 15px;
  background-color: #fff3cd;
  border: 1px solid #ffeeba;
  border-radius: 4px;
  color: #856404;
}

.alert-icon {
  font-size: 24px;
  margin-right: 15px;
  flex-shrink: 0;
}

.alert-content {
  flex-grow: 1;
}

.alert-title {
  font-weight: bold;
  margin-top: 0;
  margin-bottom: 10px;
}

.error-details {
  margin-top: 10px;
  background-color: rgba(0,0,0,0.05);
  padding: 10px;
  border-radius: 4px;
}

.error-title {
  font-weight: bold;
  margin-top: 0;
  margin-bottom: 5px;
}

.error-details ul {
  margin: 0;
  padding-left: 20px;
}

.diagnostic-tip {
  margin-top: 10px;
  margin-bottom: 5px;
}

.diagnostic-tip strong {
  color: #721c24;
}

.error-details ol {
  margin: 5px 0;
  padding-left: 20px;
}

.error-details code {
  background-color: #f8f9fa;
  padding: 2px 4px;
  border-radius: 3px;
  font-family: monospace;
}

.error-details a {
  color: #0056b3;
  text-decoration: underline;
}

.mock-data-warning {
  margin-bottom: 20px;
  padding: 15px;
  background-color: #fff3cd;
  border: 1px solid #ffeeba;
  border-radius: 4px;
  color: #856404;
}

.warning-header {
  display: flex;
  align-items: center;
  margin-bottom: 10px;
}

.warning-header h3 {
  margin: 0;
  flex-grow: 1;
}

.retry-button, .toggle-button {
  margin-left: 10px;
  padding: 5px 10px;
  border-radius: 4px;
  cursor: pointer;
}

.retry-button {
  background-color: #007bff;
  border: 1px solid #0069d9;
  color: white;
}

.toggle-button {
  background-color: #6c757d;
  border: 1px solid #5a6268;
  color: white;
}

.api-diagnostics {
  margin-top: 15px;
  font-size: 14px;
}

.api-errors {
  margin-top: 10px;
  padding: 10px;
  background-color: #f8d7da;
  border: 1px solid #f5c6cb;
  border-radius: 4px;
  color: #721c24;
}

.api-test {
  margin-top: 15px;
  padding: 10px;
  background-color: #e2e3e5;
  border: 1px solid #d6d8db;
  border-radius: 4px;
}

.api-test-input {
  display: flex;
  margin: 10px 0;
}

.api-test-input input {
  flex-grow: 1;
  padding: 5px;
  border: 1px solid #ced4da;
  border-radius: 4px 0 0 4px;
}

.api-test-input button {
  padding: 5px 10px;
  background-color: #28a745;
  border: 1px solid #28a745;
  border-radius: 0 4px 4px 0;
  color: white;
  cursor: pointer;
}

.troubleshooting {
  margin-top: 15px;
  padding: 10px;
  background-color: #d1ecf1;
  border: 1px solid #bee5eb;
  border-radius: 4px;
  color: #0c5460;
}

/* API 调试面板样式 */
.api-debug-panel {
  margin-bottom: 20px;
  padding: 15px;
  background-color: #f8f9fa;
  border: 1px solid #dee2e6;
  border-radius: 4px;
}

.api-debug-panel h4 {
  margin-top: 0;
  margin-bottom: 10px;
  color: #343a40;
}

.info-text {
  margin-bottom: 15px;
  color: #495057;
}

.api-test-controls {
  display: flex;
  flex-wrap: wrap;
  gap: 20px;
  margin-bottom: 15px;
}

.api-path-options {
  flex: 1;
  min-width: 300px;
}

.api-path-options > div {
  margin-bottom: 8px;
}

.api-path-options input[type="text"] {
  margin-left: 10px;
  padding: 4px 8px;
  width: 150px;
}

.api-test-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  align-items: flex-start;
}

.test-api-btn {
  background-color: #007bff;
  color: white;
  border: none;
  padding: 8px 16px;
  border-radius: 4px;
  cursor: pointer;
}

.view-logs-btn {
  background-color: #6c757d;
  color: white;
  border: none;
  padding: 8px 16px;
  border-radius: 4px;
  cursor: pointer;
}

.use-mock-btn {
  background-color: #17a2b8;
  color: white;
  border: none;
  padding: 8px 16px;
  border-radius: 4px;
  cursor: pointer;
}

/* API测试结果样式 */
.api-test-results {
  margin-top: 15px;
  border: 1px solid #dee2e6;
  border-radius: 4px;
  overflow: hidden;
}

.api-test-results h5 {
  margin: 0;
  padding: 10px 15px;
  background-color: #e9ecef;
  border-bottom: 1px solid #dee2e6;
}

.api-test-result {
  display: flex;
  align-items: center;
  padding: 10px 15px;
  border-bottom: 1px solid #dee2e6;
}

.api-test-result:last-child {
  border-bottom: none;
}

.result-badge {
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 0.8em;
  font-weight: bold;
  margin-right: 10px;
  min-width: 50px;
  text-align: center;
}

.result-badge.success {
  background-color: #28a745;
  color: white;
}

.result-badge.error {
  background-color: #dc3545;
  color: white;
}

.result-endpoint {
  width: 150px;
  font-weight: bold;
  margin-right: 10px;
}

.result-message {
  flex: 1;
  color: #6c757d;
}

/* API日志样式 */
.api-logs {
  margin-top: 15px;
  border: 1px solid #dee2e6;
  border-radius: 4px;
  max-height: 300px;
  overflow-y: auto;
}

.api-logs h5 {
  margin: 0;
  padding: 10px 15px;
  background-color: #e9ecef;
  border-bottom: 1px solid #dee2e6;
  position: sticky;
  top: 0;
  z-index: 1;
}

.api-log-entry {
  padding: 8px 15px;
  border-bottom: 1px solid #f1f1f1;
  font-family: monospace;
  font-size: 0.9em;
}

.log-time {
  color: #6c757d;
  margin-right: 10px;
}

.log-type {
  display: inline-block;
  margin-right: 10px;
  min-width: 50px;
  font-weight: bold;
}

.log-type.info {
  color: #007bff;
}

.log-type.error {
  color: #dc3545;
}

.log-type.warning {
  color: #ffc107;
}

.log-type.success {
  color: #28a745;
}

.log-message {
  color: #212529;
}

/* 故障排除提示样式 */
.troubleshooting-tips {
  margin-top: 15px;
  padding: 10px;
  background-color: #e2e3e5;
  border-radius: 4px;
}

.troubleshooting-tips ol {
  margin-top: 8px;
  margin-bottom: 0;
  padding-left: 25px;
}

.troubleshooting-tips li {
  margin-bottom: 5px;
}

.troubleshooting-tips code {
  background-color: #f1f3f5;
  padding: 2px 4px;
  border-radius: 3px;
  font-family: monospace;
  font-size: 0.9em;
}

.edit-form {
  margin-top: 20px;
  padding: 15px;
  background-color: #f8f9fa;
  border: 1px solid #dee2e6;
  border-radius: 4px;
}

.edit-actions {
  display: flex;
  justify-content: space-between;
  margin-top: 15px;
}

.update-btn, .cancel-btn {
  padding: 10px 20px;
  border: none;
  border-radius: 4px;
  cursor: pointer;
}

.update-btn {
  background-color: #28a745;
  color: white;
}

.cancel-btn {
  background-color: #dc3545;
  color: white;
}

.update-btn:hover {
  background-color: #218838;
}

.cancel-btn:hover {
  background-color: #c82333;
}

/* 编辑中的任务高亮显示 */
tr.editing-task {
  background-color: #e2f2fd !important;
}

.edit-form h4 {
  margin-top: 0;
  color: #28a745;
  border-bottom: 2px solid #28a745;
  padding-bottom: 10px;
  margin-bottom: 20px;
}

.promql-tag.custom-label {
  background-color: #e6f7ff;
  border: 1px solid #91d5ff;
  color: #1890ff;
}

.no-promql-selected {
  color: #6c757d;
  font-style: italic;
}

/* 发送时间选择样式 */
.weekday-selection {
  margin-top: 8px;
  width: 200px; /* 设置固定宽度 */
}

.weekday-selection select {
  width: 100%;
  padding: 8px 12px;
  border: 1px solid #ddd;
  border-radius: 4px;
  font-size: 14px;
  color: #333;
  background-color: #fff;
  cursor: pointer;
  appearance: none; /* 移除默认的下拉箭头 */
  -webkit-appearance: none;
  -moz-appearance: none;
  background-image: url("data:image/svg+xml;charset=UTF-8,%3csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 24 24' fill='none' stroke='currentColor' stroke-width='2' stroke-linecap='round' stroke-linejoin='round'%3e%3cpolyline points='6 9 12 15 18 9'%3e%3c/polyline%3e%3c/svg%3e");
  background-repeat: no-repeat;
  background-position: right 8px center;
  background-size: 16px;
  padding-right: 32px; /* 为下拉箭头留出空间 */
}

.weekday-selection select:hover {
  border-color: #999;
}

.weekday-selection select:focus {
  outline: none;
  border-color: #4CAF50;
  box-shadow: 0 0 0 2px rgba(76, 175, 80, 0.2);
}

/* 时间输入框组样式 */
.time-input-group {
  display: flex;
  gap: 10px;
  align-items: center;
  margin-top: 8px;
}

.time-input-group input[type="time"] {
  width: 200px; /* 与周几选择保持一致的宽度 */
  padding: 8px 12px;
  border: 1px solid #ddd;
  border-radius: 4px;
  font-size: 14px;
  color: #333;
}

.time-input-group input[type="time"]:hover {
  border-color: #999;
}

.time-input-group input[type="time"]:focus {
  outline: none;
  border-color: #4CAF50;
  box-shadow: 0 0 0 2px rgba(76, 175, 80, 0.2);
}

.send-times {
  margin: 1rem 0;
}

.send-time-item {
  display: flex;
  gap: 1rem;
  margin-bottom: 0.5rem;
  align-items: center;
}

.send-time-item select,
.send-time-item input {
  padding: 0.5rem;
  border: 1px solid #ddd;
  border-radius: 4px;
}

.send-time-item button {
  padding: 0.5rem 1rem;
  background-color: #ff4444;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
}

.send-time-item button:hover {
  background-color: #cc0000;
}

.task-send-times {
  margin-top: 0.5rem;
  color: #666;
}

/* 添加相关样式 */
.send-times {
    margin: 1rem 0;
}

.send-time-item {
    display: flex;
    gap: 1rem;
    margin-bottom: 0.5rem;
    align-items: center;
}

.send-time-item select,
.send-time-item input {
    padding: 0.5rem;
    border: 1px solid #ddd;
    border-radius: 4px;
    min-width: 120px;
}

.remove-time-btn {
    padding: 0.5rem 1rem;
    background-color: #dc3545;
    color: white;
    border: none;
    border-radius: 4px;
    cursor: pointer;
}

.add-time-btn {
    margin-top: 0.5rem;
    padding: 0.5rem 1rem;
    background-color: #28a745;
    color: white;
    border: none;
    border-radius: 4px;
    cursor: pointer;
}

.send-times-list {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
}

.send-time {
    font-size: 0.9em;
    color: #666;
}

.no-times {
    font-style: italic;
    color: #999;
}

.error-message {
  background-color: #ffebee;
  color: #c62828;
  padding: 10px;
  margin-bottom: 15px;
  border-radius: 4px;
  border: 1px solid #ef9a9a;
}

.no-data-message {
  text-align: center;
  padding: 20px;
  background-color: #f5f5f5;
  border-radius: 4px;
  margin: 20px 0;
}

.error-hint {
  color: #c62828;
  font-size: 0.9em;
  margin-top: 10px;
}
</style>
