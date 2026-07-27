package installer

const wizardHTML = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>Xt-PMS 安装向导</title>
<link rel="icon" href="data:image/svg+xml,<svg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 595.28 595.28'><path d='M11.39 458h109.56s6.16-295.14 104.77-280.07c0 0-66.42-47-126-.56C40.15 223.81 17.55 349.81 11.39 458zM241.48 248.46s-45.2 52.04-58.21 124.63c0 0 68.48 125.31 245.15 88.34 0 0-132.85-10.27-186.94-212.97zM170.94 132.73s117.78 4.11 183.52 212.28c0 0 55.47-75.33 58.89-125.31 0 0-65.05-119.84-242.41-86.97zM472.25 137.52h111.62s4.11 218.44-103.4 290.35c-52.04 27.39-109.56-10.27-109.56-10.27 0 0 46.57 3.42 72.59-68.48 26.02-71.9 31.5-167.09 28.76-211.6z' fill='%2304314D'/></svg>">
<style>
/* TDesign 设计令牌对齐 */
:root {
	--td-brand-color: #0052d9;
	--td-brand-color-hover: #003cab;
	--td-brand-color-light: rgba(0,82,217,.1);
	--td-text-primary: rgba(0,0,0,.9);
	--td-text-secondary: rgba(0,0,0,.6);
	--td-text-placeholder: rgba(0,0,0,.4);
	--td-border-color: #dcdcdc;
	--td-bg-container: #fff;
	--td-bg-page: #f5f5f5;
	--td-radius: 3px;
	--td-radius-medium: 6px;
	--td-size-m: 32px;
	--td-font-size-base: 14px;
	--td-success: #2ba471;
	--td-error: #d54941;
	--td-error-light: rgba(213,73,65,.1);
}

*{margin:0;padding:0;box-sizing:border-box}
body{font-family:"PingFang SC","Microsoft YaHei","Helvetica Neue",Arial,sans-serif;background:var(--td-bg-page);min-height:100vh;display:flex;flex-direction:column;align-items:center;justify-content:center}
.card{background:var(--td-bg-container);border-radius:var(--td-radius-medium);box-shadow:0 1px 4px rgba(0,0,0,.1);width:90%;max-width:580px;padding:48px 48px 44px}
h1{font-size:24px;font-weight:500;margin-bottom:16px;color:var(--td-text-primary);display:flex;align-items:center;gap:8px}
h1 svg{flex-shrink:0}
.sub{color:var(--td-text-secondary);font-size:var(--td-font-size-base);margin-bottom:40px}

/* 步骤条 */
.steps{display:flex;margin-bottom:36px;gap:0}
.step-item{flex:1;text-align:center;position:relative}
.step-item::after{content:'';position:absolute;top:14px;left:calc(50% + 20px);width:calc(100% - 40px);height:2px;background:var(--td-border-color);z-index:0}
.step-item:last-child::after{display:none}
.step-num{width:30px;height:30px;border-radius:50%;background:var(--td-border-color);color:var(--td-text-secondary);display:inline-flex;align-items:center;justify-content:center;font-size:14px;font-weight:600;position:relative;z-index:1;transition:all .2s}
.step-label{font-size:12px;color:var(--td-text-placeholder);margin-top:6px;transition:all .2s}
.step-item.active .step-num{background:var(--td-brand-color);color:#fff}
.step-item.active .step-label{color:var(--td-brand-color);font-weight:500}
.step-item.done .step-num{background:var(--td-success);color:#fff}
.step-item.done .step-label{color:var(--td-success)}
.step-item.done::after{background:var(--td-success)}

/* 步骤内容 */
.step-panel{display:none;min-height:380px}
.step-panel.active{display:flex;flex-direction:column}
.step-panel .btn-row{margin-top:auto}

/* 表单 - TDesign Input 对齐 */
.form-group{margin-bottom:24px}
.form-row{display:flex;gap:20px}
.form-row .form-group{flex:1}
label{display:block;font-size:var(--td-font-size-base);margin-bottom:10px;color:var(--td-text-primary)}
input,select{width:100%;height:var(--td-size-m);padding:0 8px;border:1px solid var(--td-border-color);border-radius:var(--td-radius);font-size:var(--td-font-size-base);font-family:inherit;color:var(--td-text-primary);background:var(--td-bg-container);outline:none;transition:all .2s;line-height:var(--td-size-m)}
input::placeholder{color:var(--td-text-placeholder)}
input:hover,select:hover{border-color:#bbb}
input:focus,select:focus{border-color:var(--td-brand-color);box-shadow:0 0 0 2px var(--td-brand-color-light)}
select{appearance:none;-webkit-appearance:none;background-image:url("data:image/svg+xml,%3Csvg width='12' height='8' viewBox='0 0 12 8' fill='none' xmlns='http://www.w3.org/2000/svg'%3E%3Cpath d='M1 1.5l5 5 5-5' stroke='%23999' stroke-width='1.5' stroke-linecap='round' stroke-linejoin='round'/%3E%3C/svg%3E");background-repeat:no-repeat;background-position:right 8px center;padding-right:28px}
.help{font-size:12px;color:var(--td-text-placeholder);margin-top:6px}

/* 按钮 - TDesign Button 对齐 */
.btn{display:inline-flex;align-items:center;justify-content:center;height:var(--td-size-m);padding:0 16px;border:none;border-radius:var(--td-radius);font-size:var(--td-font-size-base);font-weight:400;cursor:pointer;transition:all .2s;gap:4px;font-family:inherit;white-space:nowrap}
.btn-primary{background:var(--td-brand-color);color:#fff}
.btn-primary:hover{background:var(--td-brand-color-hover)}
.btn-primary:active{background:#002a7c}
.btn-primary:disabled{background:#bbd3fb;cursor:not-allowed}
.btn-success{background:var(--td-success);color:#fff}
.btn-error{background:var(--td-error);color:#fff}
.btn-default{background:var(--td-bg-container);border:1px solid var(--td-border-color);color:var(--td-text-primary)}
.btn-default:hover{border-color:var(--td-brand-color);color:var(--td-brand-color)}
.btn-default:active{border-color:var(--td-brand-color-hover);color:var(--td-brand-color-hover)}
.btn-sm{height:24px;padding:0 8px;font-size:12px;white-space:nowrap}
.btn-row{display:flex;gap:12px;margin-top:32px;justify-content:flex-end}

/* 测试结果 - TDesign Message 对齐 */
.test-msg{margin-top:8px;font-size:13px;padding:8px 12px;border-radius:var(--td-radius);line-height:20px}
.test-msg.success{background:#e8f8f2;color:var(--td-success);border:1px solid #b5e3d1}
.test-msg.error{background:var(--td-error-light);color:var(--td-error);border:1px solid #f3c2bf}

/* 进度 */
.install-progress{display:none}
.install-progress.active{display:block}
.progress-item{display:flex;align-items:center;gap:8px;padding:10px 0;font-size:var(--td-font-size-base);color:var(--td-text-placeholder)}
.progress-item.done{color:var(--td-success)}
.progress-item.running{color:var(--td-brand-color);font-weight:500}
.progress-item.fail{color:var(--td-error)}
.progress-dot{width:8px;height:8px;border-radius:50%;background:var(--td-border-color);flex-shrink:0}
.progress-item.done .progress-dot{background:var(--td-success)}
.progress-item.running .progress-dot{background:var(--td-brand-color);animation:pulse 1s infinite}
.progress-item.fail .progress-dot{background:var(--td-error)}
@keyframes pulse{0%,100%{opacity:1}50%{opacity:.4}}

/* 完成页 - TDesign Result 对齐 */
.result-wrap{display:flex;flex-direction:column;align-items:center;padding:32px 0}
.result-icon{margin-bottom:24px}
.result-icon svg{width:72px;height:72px}
.result-title{font-size:18px;font-weight:600;color:var(--td-text-primary);margin-bottom:8px;text-align:center}
.result-desc{color:var(--td-text-secondary);font-size:var(--td-font-size-base);margin-bottom:32px;text-align:center}
.result-actions{display:flex;gap:12px;justify-content:center;flex-wrap:wrap}

/* 版权信息 */
.footer{text-align:center;padding:20px;font-size:12px;color:var(--td-text-placeholder)}
</style>
</head>
<body>
<div class="card" id="app">
	<h1><svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 595.28 595.28" style="width:36px;height:36px"><path d="M11.39 458h109.56s6.16-295.14 104.77-280.07c0 0-66.42-47-126-.56C40.15 223.81 17.55 349.81 11.39 458zM241.48 248.46s-45.2 52.04-58.21 124.63c0 0 68.48 125.31 245.15 88.34 0 0-132.85-10.27-186.94-212.97zM170.94 132.73s117.78 4.11 183.52 212.28c0 0 55.47-75.33 58.89-125.31 0 0-65.05-119.84-242.41-86.97zM472.25 137.52h111.62s4.11 218.44-103.4 290.35c-52.04 27.39-109.56-10.27-109.56-10.27 0 0 46.57 3.42 72.59-68.48 26.02-71.9 31.5-167.09 28.76-211.6z" fill="#04314D"/></svg>Xt-PMS 安装向导</h1>
	<p class="sub"></p>

	<!-- 步骤指示器 -->
	<div class="steps">
		<div class="step-item active" id="step-1-indicator"><div class="step-num">1</div><div class="step-label">数据库</div></div>
		<div class="step-item" id="step-2-indicator"><div class="step-num">2</div><div class="step-label">管理员</div></div>
		<div class="step-item" id="step-3-indicator"><div class="step-num">3</div><div class="step-label">确认安装</div></div>
	</div>

	<!-- 步骤 1：数据库 + 服务配置 -->
	<div class="step-panel active" id="step-1">
		<div class="form-row"><div class="form-group"><label>主机地址</label><input id="db-host" value="127.0.0.1" placeholder="127.0.0.1"></div><div class="form-group"><label>端口</label><input id="db-port" value="3306" placeholder="3306"></div></div>
		<div class="form-row"><div class="form-group"><label>用户名</label><input id="db-user" value="root" placeholder="root"></div><div class="form-group"><label>密码</label><input id="db-password" type="password" placeholder="数据库密码"></div></div>
		<div class="form-group"><label>数据库名</label><input id="db-name" value="production_pms" placeholder="production_pms"><div class="help">确保数据库已创建（可为空库），安装向导会自动建表</div></div>
		<button class="btn btn-default" onclick="testDB()" id="btn-test">测试连接</button>
		<div style="margin-top:24px;padding-top:24px;border-top:1px solid var(--td-border-color)">
			<div class="form-row"><div class="form-group"><label>运行模式</label><select id="app-mode"><option value="production">production（生产）</option><option value="development">development（开发）</option></select></div><div class="form-group"><label>服务端口</label><input id="server-port" value="9898" placeholder="9898"></div></div>
		</div>
		<div class="btn-row"><button class="btn btn-primary" id="btn-next-1" disabled onclick="nextStep(2)">下一步 →</button></div>
	</div>

	<!-- 步骤 2：管理员 -->
	<div class="step-panel" id="step-2">
		<div class="form-row"><div class="form-group"><label>登录账号</label><input id="admin-username" value="admin" placeholder="admin"></div><div class="form-group"><label>真实姓名</label><input id="admin-realname" placeholder="管理员姓名"></div></div>
		<div class="form-row"><div class="form-group"><label>登录密码</label><input id="admin-password" type="password" placeholder="至少 6 位"></div><div class="form-group"><label>确认密码</label><input id="admin-password2" type="password" placeholder="再次输入"></div></div>
		<div class="form-row"><div class="form-group"><label>邮箱</label><input id="admin-email" placeholder="admin@example.com"></div><div class="form-group"><label>手机号</label><input id="admin-phone" placeholder="13800138000"></div></div>
		<div class="help" style="color:var(--td-text-placeholder);font-size:12px">以上信息仅用于本地系统构建，不会发送至开发者以及任何第三方</div>
		<div class="btn-row">
			<button class="btn btn-default" onclick="prevStep(1)">← 上一步</button>
			<button class="btn btn-primary" onclick="nextStep(3)">下一步 →</button>
		</div>
	</div>

	<!-- 步骤 3：确认安装 -->
	<div class="step-panel" id="step-3">
		<!-- 安装确认提示 - TDesign Alert 样式 -->
		<div id="install-confirm" style="margin-bottom:24px">
			<div style="display:flex;align-items:flex-start;gap:12px;padding:16px 20px;background:#fff;border-radius:var(--td-radius-medium);border:1px solid var(--td-brand-color-light);position:relative">
				<div style="flex-shrink:0;width:20px;height:20px;display:flex;align-items:center;justify-content:center">
					<svg viewBox="0 0 24 24" fill="none" style="width:20px;height:20px"><path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm0 15c-.55 0-1-.45-1-1v-4c0-.55.45-1 1-1s1 .45 1 1v4c0 .55-.45 1-1 1zm0-8c-.55 0-1-.45-1-1V7c0-.55.45-1 1-1s1 .45 1 1v1c0 .55-.45 1-1 1z" fill="#F5A623"/></svg>
				</div>
				<div style="flex:1">
					<div style="font-weight:600;font-size:14px;color:var(--td-text-primary);margin-bottom:8px">安装前请确认</div>
					<ul style="margin:0;padding-left:16px;color:var(--td-text-secondary);font-size:13px;line-height:1.8">
						<li>若该数据库已有数据，<span style="color:var(--td-error);font-weight:500">将被清空并覆盖</span></li>
						<li>安装过程不可中断，请确保数据库连接稳定</li>
					</ul>
				</div>
				<div style="position:absolute;left:0;top:0;bottom:0;width:4px;background:#F5A623;border-radius:var(--td-radius-medium) 0 0 var(--td-radius-medium)"></div>
			</div>
		</div>

		<div class="install-progress" id="install-progress">
			<div class="progress-item" id="prog-config"><span class="progress-dot"></span>写入配置文件…</div>
			<div class="progress-item" id="prog-schema"><span class="progress-dot"></span>创建数据表（59 张）…</div>
			<div class="progress-item" id="prog-seed"><span class="progress-dot"></span>导入种子数据…</div>
			<div class="progress-item" id="prog-admin"><span class="progress-dot"></span>创建管理员账号…</div>
			<div class="progress-item" id="prog-done"><span class="progress-dot"></span>安装完成</div>
		</div>
		<div id="done-section" style="display:none">
			<div class="result-wrap">
				<div class="result-icon">
					<svg viewBox="0 0 72 72" fill="none"><circle cx="36" cy="36" r="36" fill="#E3F9E9"/><circle cx="36" cy="36" r="27" fill="#2BA471"/><path d="M23 36l9 9 18-18" stroke="#fff" stroke-width="4" stroke-linecap="round" stroke-linejoin="round"/></svg>
				</div>
				<div class="result-title">安装完成</div>
				<div class="result-desc" id="done-msg">系统正在重启，完成后将自动跳转...</div>
				<div class="result-actions">
					<span id="done-countdown" style="color:var(--td-text-placeholder);font-size:13px">正在等待服务就绪</span>
					<button class="btn btn-primary" id="btn-go-login" onclick="goToLogin()" style="display:none">前往登录页</button>
				</div>
			</div>
		</div>
		<div id="error-section" style="display:none">
			<div class="result-wrap">
				<div class="result-icon">
					<svg viewBox="0 0 72 72" fill="none"><circle cx="36" cy="36" r="36" fill="#FDECEE"/><circle cx="36" cy="36" r="27" fill="#D54941"/><path d="M27 27l18 18M45 27L27 45" stroke="#fff" stroke-width="4" stroke-linecap="round"/></svg>
				</div>
				<div class="result-title">安装失败</div>
				<div class="result-desc" id="error-msg"></div>
				<div class="result-actions">
					<button class="btn btn-default" onclick="prevStep(2)">返回修改</button>
				</div>
			</div>
		</div>
		<div class="btn-row" id="install-btns">
			<button class="btn btn-default" onclick="prevStep(2)">← 上一步</button>
			<button class="btn btn-primary" onclick="doInstall()" id="btn-install">开始安装</button>
		</div>
	</div>
</div>

<div class="footer">© <span id="copyright-year">2024-2026</span> 广州锦杨科技有限公司</div>

<script>
let currentStep = 1;
let testPassed = false;

// 进度条映射表（全局，供 setProgress 使用）
const progressMap = {config:'prog-config',schema:'prog-schema',seed:'prog-seed',admin:'prog-admin'};

// API 路径（无需 token）
function api(path) {
	return path;
}

// === 步骤切换 ===
function showStep(n) {
	document.querySelectorAll('.step-panel').forEach(p=>p.classList.remove('active'));
	document.getElementById('step-'+n).classList.add('active');
	for(let i=1;i<=3;i++){
		let el=document.getElementById('step-'+i+'-indicator');
		el.classList.remove('active','done');
		if(i<n) el.classList.add('done');
		if(i==n) el.classList.add('active');
	}
	// 进入步骤 3 时重置安装状态
	if(n===3){
		document.getElementById('install-btns').style.display='';
		document.getElementById('install-confirm').style.display='';
		document.getElementById('install-progress').classList.remove('active');
		document.getElementById('done-section').style.display='none';
		document.getElementById('error-section').style.display='none';
		document.querySelectorAll('.progress-item').forEach(p=>p.classList.remove('done','running','fail'));
	}
	currentStep = n;
}

function nextStep(n) {
	if(n===2 && !testPassed){alert('请先测试数据库连接');return}
	if(n===3){
		let pw=document.getElementById('admin-password').value;
		let pw2=document.getElementById('admin-password2').value;
		if(pw.length<6){alert('密码至少 6 位');return}
		if(pw!==pw2){alert('两次密码不一致');return}
	}
	showStep(n);
}

function prevStep(n) { showStep(n) }

// === 数据库测试 ===
async function testDB() {
	let btn=document.getElementById('btn-test');
	btn.disabled=true; btn.textContent='测试中…';
	btn.className='btn btn-default';

	try {
		let res=await fetch(api('/api/setup/test-db'),{
			method:'POST',
			headers:{'Content-Type':'application/json'},
			body:JSON.stringify({
				host:document.getElementById('db-host').value,
				port:parseInt(document.getElementById('db-port').value)||3306,
				user:document.getElementById('db-user').value,
				password:document.getElementById('db-password').value,
				db_name:document.getElementById('db-name').value
			})
		});
		let data=await res.json();
		btn.disabled=false;
		let nextBtn=document.getElementById('btn-next-1');
		if(data.success){
			btn.className='btn btn-success';
			btn.textContent='连接成功';
			testPassed=true;
			nextBtn.disabled=false;
		}else{
			btn.className='btn btn-error';
			btn.textContent='连接失败，点击重试';
			btn.title=data.message;
			testPassed=false;
			nextBtn.disabled=true;
		}
	}catch(e){
		btn.disabled=false;
		btn.className='btn btn-error';
		btn.textContent='连接失败，点击重试';
		btn.title=e.message;
		testPassed=false;
		document.getElementById('btn-next-1').disabled=true;
	}
}

// === 生成安全配置 ===
async function genPepper() {
	let res=await fetch(api('/api/setup/generate-pepper'));
	let data=await res.json();
	document.getElementById('pepper').value=data.value;
}
async function genJWT() {
	let res=await fetch(api('/api/setup/generate-jwt'));
	let data=await res.json();
	document.getElementById('jwt-secret').value=data.value;
}

// === 安装 ===
async function doInstall() {
	// 先检查数据库是否为空
	try {
		let res=await fetch(api('/api/setup/check-db'),{
			method:'POST',
			headers:{'Content-Type':'application/json'},
			body:JSON.stringify({
				host:document.getElementById('db-host').value,
				port:parseInt(document.getElementById('db-port').value)||3306,
				user:document.getElementById('db-user').value,
				password:document.getElementById('db-password').value,
				db_name:document.getElementById('db-name').value
			})
		});
		let data=await res.json();
		if(data.success && !data.empty && !confirm(data.message+'\\n\\n确定要继续安装吗？这将删除所有现有数据！')){
			return;
		}
	}catch(e){
		alert('数据库检查失败：'+e.message);
		return;
	}

	document.getElementById('install-btns').style.display='none';
	document.getElementById('install-confirm').style.display='none';
	document.getElementById('install-progress').classList.add('active');
	let btn=document.getElementById('btn-install');
	btn.disabled=true;

	let steps=['prog-config','prog-schema','prog-seed','prog-admin','prog-done'];

	try{
		// 先标记当前步骤
		setProgress('config','running');
		let res=await fetch(api('/api/setup/install'),{
			method:'POST',
			headers:{'Content-Type':'application/json'},
			body:JSON.stringify({
				db_host:document.getElementById('db-host').value,
				db_port:parseInt(document.getElementById('db-port').value)||3306,
				db_user:document.getElementById('db-user').value,
				db_password:document.getElementById('db-password').value,
				db_name:document.getElementById('db-name').value,
				// pepper 和 jwt_secret 由后端自动生成（见 handler.go 中的默认值兜底）
				app_mode:document.getElementById('app-mode').value,
				server_port:parseInt(document.getElementById('server-port').value)||8080,
				cookie_secure:false,
				admin_username:document.getElementById('admin-username').value,
				admin_password:document.getElementById('admin-password').value,
				admin_real_name:document.getElementById('admin-realname').value,
				admin_email:document.getElementById('admin-email').value,
				admin_phone:document.getElementById('admin-phone').value
			})
		});
		let data=await res.json();
		if(!data.success){
			// 标记失败步骤
			let failStep=progressMap[data.step]||'prog-config';
			document.querySelectorAll('.progress-item').forEach(p=>p.classList.add('done'));
			document.getElementById(failStep).classList.add('fail');
			document.getElementById(failStep).classList.remove('done');
			document.getElementById('install-progress').classList.remove('active');
			document.getElementById('error-msg').textContent=data.message;
			document.getElementById('error-section').style.display='block';
			return;
		}
		// 全部完成
		document.querySelectorAll('.progress-item').forEach(p=>p.classList.add('done'));
		document.getElementById('install-progress').classList.remove('active');
		document.getElementById('done-section').style.display='block';
		autoRedirect();
	}catch(e){
		document.getElementById('install-progress').classList.remove('active');
		document.getElementById('error-msg').textContent='安装请求失败：'+e.message;
		document.getElementById('error-section').style.display='block';
	}
}

function setProgress(step, state){
	let el=document.getElementById(progressMap[step]);
	if(el){
		el.classList.remove('done','running','fail');
		if(state==='running') el.classList.add('running');
		if(state==='done') el.classList.add('done');
		if(state==='fail') el.classList.add('fail');
	}
}

function goToLogin() {
	// 使用当前访问的 origin（包含协议、主机、端口），适配反代/端口映射场景
	window.location.href = window.location.origin;
}

// autoRedirect 轮询检测服务重启完成，自动跳转登录页
async function autoRedirect() {
	let countdown = document.getElementById('done-countdown');
	let btnGo = document.getElementById('btn-go-login');
	let retries = 0;
	let maxRetries = 30; // 最多等待 30 秒
	let interval = 1000;  // 每秒检测一次

	while (retries < maxRetries) {
		countdown.textContent = '正在等待服务就绪... (' + (maxRetries - retries) + 's)';
		try {
			let res = await fetch(window.location.origin + '/api/ping', { signal: AbortSignal.timeout(2000) });
			if (res.ok) {
				countdown.textContent = '服务已就绪，即将跳转...';
				setTimeout(goToLogin, 500);
				return;
			}
		} catch(e) {
			// 服务尚未就绪，继续等待
		}
		retries++;
		await new Promise(r => setTimeout(r, interval));
	}
	// 超时则显示手动跳转按钮
	countdown.style.display = 'none';
	btnGo.style.display = '';
}

// === 初始化 ===
var y = new Date().getFullYear();
document.getElementById('copyright-year').textContent = y <= 2024 ? '2024' : '2024 - ' + y;
// pepper 和 jwt-secret 由后端自动生成，无需前端预填
</script>
</body>
</html>
`
