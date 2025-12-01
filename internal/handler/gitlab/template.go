package gitlab

// GetJobListTemplate 返回GitLab流水线任务列表页面的HTML模板
func GetJobListTemplate() string {
	return `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>女娲</title>
    <style>
        body {
            font-family: Arial, sans-serif;
            margin: 0;
            background-color: #f5f5f5;
        }
        .container {
            width: 100%;
            max-width: 100%;
            margin: 0;
            background-color: white;
            padding: 20px;
            box-sizing: border-box;
            border-radius: 0;
            box-shadow: none;
        }
        h1 {
            color: #333;
            border-bottom: 2px solid #eee;
            padding-bottom: 10px;
        }
        table {
            width: 100%;
            border-collapse: collapse;
            margin-top: 20px;
        }
        th, td {
            padding: 12px;
            text-align: left;
            border-bottom: 1px solid #ddd;
        }
        th {
            background-color: #f8f9fa;
            font-weight: bold;
            color: #555;
        }
        tr:hover {
            background-color: #f5f5f5;
        }
        .status {
            padding: 4px 8px;
            border-radius: 4px;
            font-size: 12px;
            font-weight: bold;
            text-transform: uppercase;
        }
        .status-success {
            background-color: #d4edda;
            color: #155724;
        }
        .status-failed {
            background-color: #f8d7da;
            color: #721c24;
        }
        .status-running {
            background-color: #cce7ff;
            color: #004085;
        }
        .status-pending {
            background-color: #fff3cd;
            color: #856404;
        }
        .commit-id {
            font-family: monospace;
            font-size: 12px;
        }
        .actions {
            text-align: right;
        }
        .btn {
            padding: 6px 12px;
            border: none;
            border-radius: 4px;
            cursor: pointer;
            text-decoration: none;
            font-size: 12px;
        }
        .btn-primary {
            background-color: #007bff;
            color: white;
        }
        .btn-secondary {
            background-color: #6c757d;
            color: white;
        }
        .filter-bar {
            margin-bottom: 20px;
            display: flex;
            gap: 10px;
        }
        .filter-bar input, .filter-bar select {
            padding: 8px;
            border: 1px solid #ddd;
            border-radius: 4px;
        }
        .filter-bar button {
            padding: 8px 16px;
            background-color: #28a745;
            color: white;
            border: none;
            border-radius: 4px;
            cursor: pointer;
        }
        .actions-bar { margin: 12px 0 20px 0; display: flex; gap: 12px; flex-wrap: wrap; }
        .action-group { display: flex; align-items: center; gap: 8px; padding: 10px; border: 1px solid #eee; border-radius: 6px; background-color: #f8f9fa; }
        .merge-arrow { font-weight: bold; color: #2c3e50; padding: 0 6px; }
        .tag { display:inline-block; padding: 2px 8px; border-radius: 12px; font-size: 12px; line-height: 16px; }
        .tag-create { background:#e8f0fe; color:#1a73e8; border:1px solid #c6dcff; }
        .tag-merge { background:#fff4e5; color:#b4690e; border:1px solid #ffd8a8; }
        .tag-edit { background:#f1f3f4; color:#5f6368; border:1px solid #e0e0e0; }
    </style>
</head>
<body>
    <div class="container">
        <h1>女娲</h1>
        <div class="actions-bar">
            <div class="action-group">
                <span>创建分支</span>
                <select id="branchPrefix" style="width:140px">
                    <option value="feature/">feature/</option>
                    <option value="test/">test/</option>
                    <option value="release/">release/</option>
                </select>
                <input id="branchSubject" type="text" placeholder="主体名称" style="width:160px">
                <button class="btn btn-primary" onclick="createBranch()">创建分支</button>
            </div>
            <div class="action-group">
                <span>分支合并</span>
                <select id="mrSource" class="select-sm" style="width:160px"></select>
                <span class="merge-arrow">→</span>
                <select id="mrTarget" class="select-sm" style="width:160px"></select>
                <input id="mrTitle" type="text" placeholder="MR标题" style="width:200px">
                <input id="mrDesc" type="text" placeholder="描述(可选)" style="width:200px">
                <label><input type="checkbox" id="mrSquash">Squash</label>
                <label><input type="checkbox" id="mrRemoveSource">移除源分支</label>
                <label><input type="checkbox" id="mrMWPS">等待流水线成功</label>
                <button class="btn btn-primary" onclick="createMergeRequest()">创建MR</button>
                <input id="mrIID" type="number" placeholder="MR编号" style="width:100px">
                <input id="mrMessage" type="text" placeholder="合并提交信息(可选)" style="width:200px">
                <button class="btn btn-primary" onclick="acceptMergeRequest()">接受MR</button>
                <button class="btn btn-primary" onclick="autoMergeRequest()">自动更新分支</button>
            </div>
            <div class="action-group" style="display:none">
                <span>阶段推进</span>
                <select id="promotePrefix" style="width:120px">
                    <option value="feature">feature</option>
                    <option value="test">test</option>
                    <option value="release">release</option>
                </select>
                <select id="promoteName" style="width:160px"></select>
                <input id="promoteTitle" type="text" placeholder="标题(可选)" style="width:180px">
                <input id="promoteDesc" type="text" placeholder="描述(可选)" style="width:200px">
                <input id="promoteMessage" type="text" placeholder="提交信息" style="width:260px">
                <label><input type="checkbox" id="promoteSquash">Squash</label>
                <label><input type="checkbox" id="promoteRemoveSource">移除源分支</label>
                <label><input type="checkbox" id="promoteMWPS">等待流水线成功</label>
                <button class="btn btn-primary" onclick="promoteStage()">推进</button>
            </div>
        </div>
        <div id="actionMsg" style="margin:8px 0; color:#333;"></div>
        
        <div class="filter-bar">
            <select id="branchFilter" onchange="filterJobs()"><option value="">所有分支</option></select>
            <select id="environmentFilter" onchange="filterJobs()"><option value="">所有环境</option></select>
            <select id="statusFilter" onchange="filterJobs()">
                <option value="">所有状态</option>
                <option value="success">成功</option>
                <option value="failed">失败</option>
                <option value="running">运行中</option>
                <option value="pending">待处理</option>
            </select>
            <select id="triggerFilter" onchange="filterJobs()"><option value="">所有触发用户</option></select>
            <select id="authorFilter" onchange="filterJobs()"><option value="">所有提交作者</option></select>
            <input type="date" id="createdDate" onchange="filterJobs()" />
            <button onclick="refreshJobs()">刷新</button>
        </div>
        
        <table id="jobsTable">
            <thead>
                <tr>
                    <th>ID</th>
                    <th>状态</th>
                    <th>任务类型</th>
                    <th>分支</th>
                    <th>环境</th>
                    <th>触发用户</th>
                    <th>提交ID</th>
                    <th>提交信息</th>
                    <th>提交作者</th>
                    <th>创建时间</th>
                    <th>持续时间</th>
                    <th>操作</th>
                </tr>
            </thead>
            <tbody id="jobsTableBody">
                <!-- 动态填充数据 -->
            </tbody>
        </table>
        <div class="pager" style="display:flex; align-items:center; gap:10px; margin-top:12px; justify-content:flex-end;">
            <button class="btn btn-secondary" id="prevPageBtn">上一页</button>
            <span id="pageInfo"></span>
            <input id="jumpInput" type="number" min="1" placeholder="页码" style="width:80px; padding:6px; border:1px solid #ddd; border-radius:4px;">
            <button class="btn btn-secondary" id="jumpBtn">跳转</button>
            <button class="btn btn-secondary" id="nextPageBtn">下一页</button>
        </div>
    </div>

    <script>
        var currentPage = 1;
        var pageSize = 20;
        var totalPages = 1;
        var jobsController = null;
        window.onload = function() {
            bindPager();
            loadJobs();
            loadBranches();
            loadEnvironments();
            var pfEl = document.getElementById('promotePrefix');
            if(pfEl){ pfEl.onchange = refreshPromoteNameOptions; }
        };

        // 获取任务列表
        function loadJobs() {
            if (jobsController) { try { jobsController.abort(); } catch(e){} }
            jobsController = new AbortController();
            fetch('/api/gitlab/jobs?page='+currentPage+'&per_page='+pageSize, { signal: jobsController.signal })
                .then(response => {
                    if (!response.ok) {
                        throw new Error('Network response was not ok: ' + response.status);
                    }
                    return response.json();
                })
                .then(data => {
                    console.log('API response:', data); // 添加日志以便调试
                    if (data.code === 0) {
                        var payload = data.data;
                        var items = [];
                        if (Array.isArray(payload)) {
                            items = payload;
                            totalPages = 1;
                        } else {
                            items = (payload && payload.items) || [];
                            var pg = (payload && payload.pagination) || {};
                            totalPages = pg.total_pages || 1;
                            currentPage = pg.current_page || currentPage;
                        }
                        renderJobs(items);
                    } else {
                        console.error('获取任务列表失败:', data.message);
                    }
                })
                .catch(error => {
                    console.error('获取任务列表错误:', error);
                });
        }

        // 渲染任务列表
        function renderJobs(jobs) {
            console.log('Rendering jobs:', jobs); // 添加日志以便调试
            const tbody = document.getElementById('jobsTableBody');
            const frag = document.createDocumentFragment();
            tbody.innerHTML = '';

            

            jobs.forEach(job => {
                const row = document.createElement('tr');
                
                // 根据状态设置CSS类
                let statusClass = 'status-pending';
                if (job.status === 'success') {
                    statusClass = 'status-success';
                } else if (job.status === 'failed') {
                    statusClass = 'status-failed';
                } else if (job.status === 'running') {
                    statusClass = 'status-running';
                }
                
                var t = (job.task_type || '修改文件');
                var tClass = (t==='创建分支') ? 'tag-create' : ((t==='分支合并') ? 'tag-merge' : 'tag-edit');
                var typeLabel = '<span class="tag '+tClass+'">'+t+'</span>';
                var commitDisplay = (t==='创建分支') ? '创建分支' : (job.commit_message || '');

                // 构建行内容
                let rowContent = '<td>' + (job.id || job.ID || '') + '</td>';
                rowContent += '<td><span class="status ' + statusClass + '">' + job.status + '</span></td>';
                rowContent += '<td>' + typeLabel + '</td>';
                rowContent += '<td>' + (job.branch_name || job.branch || '') + '</td>';
                rowContent += '<td>' + (job.environment_name || '') + '</td>';
                rowContent += '<td>' + (job.trigger_user || '') + '</td>';
                rowContent += '<td class="commit-id">' + (job.commit_id ? job.commit_id.substring(0, 8) : '') + '</td>';
                rowContent += '<td>' + commitDisplay + '</td>';
                rowContent += '<td>' + (job.commit_author || '') + '</td>';
                rowContent += '<td>' + (job.created_at || '') + '</td>';
                rowContent += '<td>' + (job.duration || '') + '</td>';
                rowContent += '<td class="actions"><a href="' + (job.web_url || job.external_web_url || '#') + '" target="_blank" class="btn btn-primary">查看</a></td>';
                
                row.innerHTML = rowContent;
                frag.appendChild(row);
            });
            tbody.appendChild(frag);
            updateUserAuthorFilters(jobs);
            
            filterJobs();
            updatePager(jobs.length);
        }

        // 筛选任务
        function filterJobs() {
            const branchFilter = document.getElementById('branchFilter').value.toLowerCase();
            const environmentFilter = document.getElementById('environmentFilter').value.toLowerCase();
            const statusFilter = document.getElementById('statusFilter').value.toLowerCase();
            const triggerFilter = document.getElementById('triggerFilter').value.toLowerCase();
            const authorFilter = document.getElementById('authorFilter').value.toLowerCase();
            const createdDateVal = document.getElementById('createdDate').value;

            const rows = document.querySelectorAll('#jobsTableBody tr');

            rows.forEach(row => {
                const branch = row.cells[3].textContent.toLowerCase().trim();
                const environment = row.cells[4].textContent.toLowerCase().trim();
                const status = row.cells[1].textContent.toLowerCase().trim();
                const trigger = row.cells[5].textContent.toLowerCase().trim();
                const author = row.cells[8].textContent.toLowerCase().trim();
                const createdAtText = row.cells[9].textContent.trim();
                const createdAtDate = parseCreatedAt(createdAtText);
                const createdAtYMD = createdAtDate ? formatYMD(createdAtDate) : '';

                const branchMatch = branchFilter === '' || branch === branchFilter;
                const environmentMatch = environmentFilter === '' || environment === environmentFilter;
                const statusMatch = statusFilter === '' || status === statusFilter;
                const triggerMatch = triggerFilter === '' || trigger === triggerFilter;
                const authorMatch = authorFilter === '' || author === authorFilter;
                const dateMatch = (createdDateVal === '' || createdAtYMD === createdDateVal);

                if (branchMatch && environmentMatch && statusMatch && triggerMatch && authorMatch && dateMatch) {
                    row.style.display = '';
                } else {
                    row.style.display = 'none';
                }
            });
        }

        // 刷新任务列表
        function refreshJobs() { loadJobs(); }

        function bindPager(){
            document.getElementById('prevPageBtn').onclick=function(){ if(currentPage>1){ currentPage--; loadJobs(); } };
            document.getElementById('nextPageBtn').onclick=function(){ if(currentPage<totalPages){ currentPage++; loadJobs(); } };
            document.getElementById('jumpBtn').onclick=function(){ var v=parseInt(document.getElementById('jumpInput').value,10); if(!isNaN(v)&&v>=1&&v<=totalPages){ currentPage=v; loadJobs(); } };
        }

        function updatePager(returnedCount){
            var info=document.getElementById('pageInfo');
            info.textContent='第 '+currentPage+' / 共 '+totalPages+' 页';
            var prev=document.getElementById('prevPageBtn');
            var next=document.getElementById('nextPageBtn');
            prev.disabled = currentPage<=1;
            next.disabled = currentPage>=totalPages || returnedCount===0;
        }

        function loadBranches(){
            fetch('/api/gitlab/branches').then(function(r){ return r.json(); }).then(function(d){
                if(d.code!==0) return;
                var list = Array.isArray(d.data) ? d.data : ((d.data&&d.data.items)||[]);
                window._branchesList = list;
                var bf = document.getElementById('branchFilter');
                var selected = bf.value;
                bf.innerHTML = '<option value="">所有分支</option>';
                list.forEach(function(b){ var name = b.name || b.Branch || b.branch || ''; if(name){ var o=document.createElement('option'); o.value=name; o.textContent=name; bf.appendChild(o);} });
                if(selected) bf.value = selected;
                if(window._selectBranchAfterCreate){ bf.value = window._selectBranchAfterCreate; }

                var srcSel = document.getElementById('mrSource');
                var tgtSel = document.getElementById('mrTarget');
                var srcSelVal = srcSel ? srcSel.value : '';
                var tgtSelVal = tgtSel ? tgtSel.value : '';
                if(srcSel){ srcSel.innerHTML = ''; }
                if(tgtSel){ tgtSel.innerHTML = ''; }
                list.forEach(function(b){ var name = b.name || b.Branch || b.branch || ''; if(name){ if(srcSel){ var o2=document.createElement('option'); o2.value=name; o2.textContent=name; srcSel.appendChild(o2);} if(tgtSel){ var o3=document.createElement('option'); o3.value=name; o3.textContent=name; tgtSel.appendChild(o3);} } });
                if(srcSel && srcSelVal) srcSel.value = srcSelVal;
                if(tgtSel && tgtSelVal) tgtSel.value = tgtSelVal;
                refreshPromoteNameOptions();
                if(window._selectBranchAfterCreate){ try { filterJobs(); } catch(e){} window._selectBranchAfterCreate = ''; }
            }).catch(function(e){ console.error('加载分支失败', e); });
        }

        function loadEnvironments(){
            fetch('/api/environments').then(function(r){ return r.json(); }).then(function(d){
                if(d.code!==0) return;
                var list = (d.data&&d.data.items) || [];
                var ef = document.getElementById('environmentFilter');
                var selected = ef.value;
                ef.innerHTML = '<option value="">所有环境</option>';
                list.forEach(function(e){ var name = e.name || e.env_name || ''; if(name){ var o=document.createElement('option'); o.value=name; o.textContent=name; ef.appendChild(o);} });
                if(selected) ef.value = selected;
            }).catch(function(e){ console.error('加载环境失败', e); });
        }

        function updateUserAuthorFilters(jobs){
            var tf = document.getElementById('triggerFilter');
            var af = document.getElementById('authorFilter');
            var triggers = {};
            var authors = {};
            jobs.forEach(function(j){ var t=(j.trigger_user||'').trim(); var a=(j.commit_author||'').trim(); if(t){triggers[t]=true;} if(a){authors[a]=true;} });
            var tOpts = Object.keys(triggers).sort();
            var aOpts = Object.keys(authors).sort();
            var tSel = tf.value; var aSel = af.value;
            tf.innerHTML = '<option value="">所有触发用户</option>' + tOpts.map(function(v){ return '<option value="'+v+'">'+v+'</option>'; }).join('');
            af.innerHTML = '<option value="">所有提交作者</option>' + aOpts.map(function(v){ return '<option value="'+v+'">'+v+'</option>'; }).join('');
            if(tSel) tf.value = tSel; if(aSel) af.value = aSel;
        }

        function parseCreatedAt(s){
            if(!s) return null;
            var parts = s.split(' ');
            if(parts.length<2) return null;
            var d = parts[0].split('-');
            var t = parts[1].split(':');
            if(d.length<3||t.length<2) return null;
            var year = parseInt(d[0],10);
            var month = parseInt(d[1],10)-1;
            var day = parseInt(d[2],10);
            var hour = parseInt(t[0],10);
            var minute = parseInt(t[1],10);
            var second = t.length>2 ? parseInt(t[2],10) : 0;
            return new Date(year, month, day, hour, minute, second);
        }

        function formatYMD(d){
            var y = d.getFullYear();
            var m = (d.getMonth()+1).toString().padStart(2,'0');
            var day = d.getDate().toString().padStart(2,'0');
            return y+'-'+m+'-'+day;
        }

        function createBranch(){
            var pEl = document.getElementById('branchPrefix');
            var sEl = document.getElementById('branchSubject');
            var p = pEl ? (pEl.value||'') : '';
            var s = sEl ? (sEl.value||'').trim() : '';
            if(!p || !s) return;
            var name = p + s;
            var ref = 'main';
            window._selectBranchAfterCreate = name;
            fetch('/api/gitlab/branches', { method:'POST', headers:{'Content-Type':'application/json'}, body: JSON.stringify({ name:name, ref:ref }) })
                .then(function(r){ return r.json(); })
                .then(function(d){ if(d.code===0){ loadBranches(); refreshJobs(); } })
                .catch(function(e){ console.error('创建分支错误', e); });
        }

        function createMergeRequest(){
            var srcEl = document.getElementById('mrSource');
            var tgtEl = document.getElementById('mrTarget');
            var titleEl = document.getElementById('mrTitle');
            var descEl = document.getElementById('mrDesc');
            var source = srcEl ? srcEl.value : '';
            var target = tgtEl ? tgtEl.value : '';
            var title = titleEl ? titleEl.value.trim() : '';
            var desc = descEl ? descEl.value.trim() : '';
            var squash = (document.getElementById('mrSquash')||{}).checked || false;
            var remove = (document.getElementById('mrRemoveSource')||{}).checked || false;
            var mwps = (document.getElementById('mrMWPS')||{}).checked || false;
            if(!source || !target || !title) return;
            var payload = { source_branch: source, target_branch: target, title: title, description: desc, squash: squash, remove_source_branch: remove, merge_when_pipeline_succeeds: mwps };
            fetch('/api/gitlab/merge_requests', { method:'POST', headers:{'Content-Type':'application/json'}, body: JSON.stringify(payload) })
                .then(function(r){ return r.json(); })
                .then(function(d){ var msgEl=document.getElementById('actionMsg'); if(d.code===0){ msgEl.textContent='创建MR成功'; refreshJobs(); } else { msgEl.textContent='创建MR失败：'+(d.message||''); } })
                .catch(function(e){ console.error('创建MR错误', e); });
        }

        function acceptMergeRequest(){
            var iidEl = document.getElementById('mrIID');
            var msgEl = document.getElementById('mrMessage');
            var iid = iidEl ? parseInt(iidEl.value, 10) : 0;
            if(!iid || iid<=0) return;
            var payload = { squash: (document.getElementById('mrSquash')||{}).checked || false, remove_source_branch: (document.getElementById('mrRemoveSource')||{}).checked || false, merge_when_pipeline_succeeds: (document.getElementById('mrMWPS')||{}).checked || false, merge_commit_message: msgEl ? msgEl.value.trim() : '' };
            fetch('/api/gitlab/merge_requests/'+iid+'/merge', { method:'POST', headers:{'Content-Type':'application/json'}, body: JSON.stringify(payload) })
                .then(function(r){ return r.json(); })
                .then(function(d){ var msgEl=document.getElementById('actionMsg'); if(d.code===0){ msgEl.textContent='接受MR成功（可能等待流水线或审批）'; refreshJobs(); } else { msgEl.textContent='接受MR失败：'+(d.message||''); } })
                .catch(function(e){ console.error('接受MR错误', e); });
        }

        function autoMergeRequest(){
            var srcEl = document.getElementById('mrSource');
            var tgtEl = document.getElementById('mrTarget');
            var titleEl = document.getElementById('mrTitle');
            var descEl = document.getElementById('mrDesc');
            var msgEl = document.getElementById('mrMessage');
            var source = srcEl ? srcEl.value : '';
            var target = tgtEl ? tgtEl.value : '';
            var title = titleEl ? titleEl.value.trim() : '';
            var desc = descEl ? descEl.value.trim() : '';
            var message = msgEl ? msgEl.value.trim() : '';
            var squash = (document.getElementById('mrSquash')||{}).checked || false;
            var remove = (document.getElementById('mrRemoveSource')||{}).checked || false;
            var mwps = (document.getElementById('mrMWPS')||{}).checked || false;
            if(!source || !target || !title) return;
            var payload = { source_branch: source, target_branch: target, title: title, description: desc, squash: squash, remove_source_branch: remove, merge_when_pipeline_succeeds: mwps, merge_commit_message: message };
            fetch('/api/gitlab/merge_requests/auto', { method:'POST', headers:{'Content-Type':'application/json'}, body: JSON.stringify(payload) })
                .then(function(r){ return r.json(); })
                .then(function(d){ var msgEl=document.getElementById('actionMsg'); if(d.code===0){ msgEl.textContent='自动合并成功'; refreshJobs(); } else { msgEl.textContent='自动合并失败，请手动完成合并：'+(d.message||''); } })
                .catch(function(e){ console.error('自动合并错误', e); });
        }

        function refreshPromoteNameOptions(){
            var list = window._branchesList || [];
            var pfEl = document.getElementById('promotePrefix');
            var nmEl = document.getElementById('promoteName');
            if(!nmEl) return;
            var pf = pfEl ? (pfEl.value||'') : '';
            var subjects = {};
            list.forEach(function(b){
                var name = b.name || b.Branch || b.branch || '';
                if(!name) return;
                var pre = pf + '/';
                if(name.indexOf(pre) === 0){
                    var s = name.substring(pre.length);
                    if(s){ subjects[s] = true; }
                }
            });
            var opts = Object.keys(subjects).sort();
            var sel = nmEl.value;
            nmEl.innerHTML = '';
            opts.forEach(function(s){ var o=document.createElement('option'); o.value=s; o.textContent=s; nmEl.appendChild(o); });
            if(sel) nmEl.value = sel;
        }

        

        function promoteStage(){
            var spEl = document.getElementById('promotePrefix');
            var nEl = document.getElementById('promoteName');
            var tEl = document.getElementById('promoteTarget');
            var tiEl = document.getElementById('promoteTitle');
            var dEl = document.getElementById('promoteDesc');
            var sp = spEl ? (spEl.value||'') : '';
            var n = nEl ? (nEl.value||'').trim() : '';
            var t = tEl ? (tEl.value||'').trim() : '';
            var ti = tiEl ? (tiEl.value||'').trim() : '';
            var d = dEl ? (dEl.value||'').trim() : '';
            var sq = (document.getElementById('promoteSquash')||{}).checked || false;
            var rm = (document.getElementById('promoteRemoveSource')||{}).checked || false;
            var mw = (document.getElementById('promoteMWPS')||{}).checked || false;
            if(!sp || !n) return;
            var payload = { source_prefix: sp, name: n, target: t||'auto', title: ti, description: d, squash: sq, remove_source_branch: rm, merge_when_pipeline_succeeds: mw };
            fetch('/api/gitlab/promote', { method:'POST', headers:{'Content-Type':'application/json'}, body: JSON.stringify(payload) })
                .then(function(r){ return r.json(); })
                .then(function(d){ if(d.code===0){ refreshJobs(); loadBranches(); } })
                .catch(function(e){ console.error('阶段推进错误', e); });
        }
    </script>
</body>
</html>
`
}
