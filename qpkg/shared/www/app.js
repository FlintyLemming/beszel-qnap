(() => {
  const I18N = {
    zh: {
      tab_overview: "概览与状态",
      tab_config: "配置信息",
      tab_updates: "二进制与更新",
      tab_logs: "运行日志",
      setup_title: "初始化引导",
      svc_control: "运行状态与控制",
      btn_start: "启动",
      btn_restart: "重启",
      btn_stop: "停止",
      k_state: "运行状态",
      k_pid: "进程 PID",
      k_uptime: "运行时长",
      k_mem: "内存占用 (RSS)",
      k_health: "健康检查",
      k_configured: "配置就绪",
      bin_summary: "Agent 二进制信息",
      btn_manage_bin: "管理 / 更新",
      k_bin_ver: "当前版本",
      k_bin_src: "安装来源",
      k_bin_size: "文件大小",
      k_updated_at: "更新时间",
      k_fingerprint: "Agent 指纹 (Fingerprint)",
      btn_refresh: "刷新",
      btn_fp_reset: "重置指纹",
      host_diag: "QNAP 主机环境探测",
      k_nas_model: "NAS 型号 / QTS",
      k_arch: "CPU 架构",
      k_qpkg_ver: "QPKG 运行环境版本",
      k_sensors: "Hwmon 传感器",
      k_volumes: "检测到的存储卷 (可在配置页一键填入 FILESYSTEM / EXTRA_FILESYSTEMS)",
      cfg_hint: "修改配置后可点击“保存并重启”使之立即生效。",
      btn_reload: "重新加载",
      btn_save: "仅保存",
      btn_save_restart: "保存并重启 Agent",
      cfg_sec_conn: "连接与认证",
      hint_listen: "SSH 监听端口或地址，默认 45876",
      hint_sysname: "可选，覆盖上报给 Hub 的系统名称",
      hint_key: "从 Beszel Hub 添加系统时获取的公钥（支持多行多个公钥）",
      lbl_disable_ssh: "DISABLE_SSH (仅使用 WebSocket 主动连接 Hub)",
      cfg_sec_storage: "磁盘与文件系统 (QNAP 存储卷)",
      quick_fill_vol: "快速填入存储卷:",
      hint_fs: "QNAP 的 / 为内存盘，建议设置为数据卷路径如 /share/CACHEDEV1_DATA",
      hint_extra_fs: "额外监控的磁盘或挂载点，逗号分隔",
      hint_disk_cache: "磁盘用量缓存时间，避免频繁唤醒休眠硬盘",
      cfg_sec_docker: "Docker / Container Station",
      btn_detect: "填入默认",
      cfg_sec_hw: "SMART、温度传感器与 GPU",
      hint_smart_devices: "QTS 下 SATA 盘默认识别为 SCSI，建议显式指定 :sat（如 /dev/sda:sat）",
      cfg_sec_sys: "系统、运行时与额外参数",
      lbl_auto_start: "随 QPKG / NAS 启动自动运行 beszel-agent",
      cfg_sec_custom: "自定义环境变量 (兼容未来 Beszel 新版本参数)",
      btn_add_env: "+ 添加变量",
      upd_current: "当前二进制状态",
      btn_rollback: "回滚至上一版本",
      k_backup: "可用备份 (.bak)",
      upd_manual: "手动上传更新",
      upd_manual_desc: "支持直接上传 GitHub Release 的 .tar.gz / .zip 压缩包或裸 beszel-agent Linux 二进制文件（自动校验架构并备份旧版）。",
      drop_hint: "点击选择或拖拽文件到此处",
      btn_upload_install: "上传并安装",
      job_title: "在线下载任务",
      upd_online: "在线下载 / 版本管理 (henrygd/beszel)",
      btn_fetch_releases: "获取可用版本",
      lbl_dl_source: "下载源",
      opt_custom_mirror: "自定义 GitHub 加速前缀",
      lbl_custom_prefix: "自定义加速前缀",
      lbl_prefer_glibc: "优先下载 glibc 版本 (amd64 NVML GPU 需求)",
      btn_dl_url: "从此 URL 下载安装",
      releases_empty: "点击“获取可用版本”加载上游 henrygd/beszel 发布列表。",
      lbl_auto_refresh: "自动刷新",
      btn_clear: "清空",
      status_running: "运行中",
      status_stopped: "已停止",
      status_uninstalled: "未安装二进制",
      status_unconfigured: "未配置",
      yes: "是",
      no: "否",
      not_installed: "未安装",
      step_bin: "1. 安装 beszel-agent 二进制文件（在线下载或手动上传）",
      step_cfg: "2. 配置连接信息（填写 KEY 或 HUB_URL + TOKEN）",
      step_run: "3. 启动 beszel-agent 服务",
      btn_go_setup: "前往设置",
      install_ver: "下载并安装",
      current_tag: "当前版本",
      prerelease: "预发布",
      latest_tag: "最新",
      confirm_fp_reset: "确定要重置 Agent 指纹吗？重置后下次连接 Hub 会生成新指纹。",
      confirm_rollback: "确定回滚到上一版本二进制并重启吗？",
      confirm_clear_log: "确定清空当前日志文件吗？",
      auth_required: "QTS 会话未认证或已过期，请从 QTS 桌面重新打开此应用。"
    },
    en: {
      tab_overview: "Overview",
      tab_config: "Configuration",
      tab_updates: "Binary & Updates",
      tab_logs: "Logs",
      setup_title: "Getting Started",
      svc_control: "Service Status & Control",
      btn_start: "Start",
      btn_restart: "Restart",
      btn_stop: "Stop",
      k_state: "State",
      k_pid: "PID",
      k_uptime: "Uptime",
      k_mem: "Memory (RSS)",
      k_health: "Health Check",
      k_configured: "Configured",
      bin_summary: "Agent Binary",
      btn_manage_bin: "Manage / Update",
      k_bin_ver: "Installed Version",
      k_bin_src: "Source",
      k_bin_size: "File Size",
      k_updated_at: "Updated At",
      k_fingerprint: "Agent Fingerprint",
      btn_refresh: "Refresh",
      btn_fp_reset: "Reset Fingerprint",
      host_diag: "QNAP Host Environment",
      k_nas_model: "NAS Model / QTS",
      k_arch: "Architecture",
      k_qpkg_ver: "QPKG Runtime Version",
      k_sensors: "Hwmon Sensors",
      k_volumes: "Detected Storage Volumes (click in Config tab to fill FILESYSTEM / EXTRA_FILESYSTEMS)",
      cfg_hint: "Save changes, or Save & Restart to apply immediately.",
      btn_reload: "Reload",
      btn_save: "Save",
      btn_save_restart: "Save & Restart",
      cfg_sec_conn: "Connection & Auth",
      hint_listen: "SSH listen port or address (default 45876)",
      hint_sysname: "Optional override for system hostname in Hub",
      hint_key: "SSH public key(s) from Beszel Hub (supports multiple lines)",
      lbl_disable_ssh: "DISABLE_SSH (WebSocket outbound mode only)",
      cfg_sec_storage: "Storage & Filesystems (QNAP Volumes)",
      quick_fill_vol: "Quick-fill volume:",
      hint_fs: "QNAP / is a ramdisk; set this to a data volume like /share/CACHEDEV1_DATA",
      hint_extra_fs: "Comma-separated extra disks or mountpoints",
      hint_disk_cache: "Cache disk usage checks to avoid waking sleeping HDDs",
      cfg_sec_docker: "Docker / Container Station",
      btn_detect: "Fill Default",
      cfg_sec_hw: "SMART, Sensors & GPU",
      hint_smart_devices: "QTS often exposes SATA drives as SCSI; append :sat (e.g. /dev/sda:sat)",
      cfg_sec_sys: "System, Runtime & Extra Flags",
      lbl_auto_start: "Automatically start beszel-agent when QPKG / NAS boots",
      cfg_sec_custom: "Custom Environment Variables (Forward-compatible with future Beszel releases)",
      btn_add_env: "+ Add Variable",
      upd_current: "Current Binary",
      btn_rollback: "Rollback to Backup",
      k_backup: "Backup (.bak)",
      upd_manual: "Manual Upload",
      upd_manual_desc: "Upload a GitHub Release .tar.gz / .zip archive or a raw beszel-agent Linux ELF binary.",
      drop_hint: "Click to choose or drag & drop file here",
      btn_upload_install: "Upload & Install",
      job_title: "Online Download Task",
      upd_online: "Online Download / Versions (henrygd/beszel)",
      btn_fetch_releases: "Fetch Releases",
      lbl_dl_source: "Download Source",
      opt_custom_mirror: "Custom GitHub Mirror Prefix",
      lbl_custom_prefix: "Custom Mirror Prefix",
      lbl_prefer_glibc: "Prefer glibc build (amd64 NVML GPU support)",
      btn_dl_url: "Download & Install URL",
      releases_empty: "Click 'Fetch Releases' to load available versions from henrygd/beszel.",
      lbl_auto_refresh: "Auto-refresh",
      btn_clear: "Clear",
      status_running: "Running",
      status_stopped: "Stopped",
      status_uninstalled: "Binary Missing",
      status_unconfigured: "Not Configured",
      yes: "Yes",
      no: "No",
      not_installed: "Not installed",
      step_bin: "1. Install beszel-agent binary (download online or upload manually)",
      step_cfg: "2. Configure connection (KEY or HUB_URL + TOKEN)",
      step_run: "3. Start beszel-agent service",
      btn_go_setup: "Configure",
      install_ver: "Download & Install",
      current_tag: "Installed",
      prerelease: "Pre-release",
      latest_tag: "Latest",
      confirm_fp_reset: "Reset agent fingerprint? A new fingerprint will be generated on next connection.",
      confirm_rollback: "Roll back to the previous backup binary and restart?",
      confirm_clear_log: "Clear the current log file?",
      auth_required: "QTS session missing or expired. Please reopen this app from QTS Desktop."
    }
  };

  let lang = localStorage.getItem("beszel_qnap_lang") || (navigator.language && navigator.language.startsWith("zh") ? "zh" : "en");
  let latestStatus = null;
  let currentLogKind = "agent";
  let logTimer = null;
  let jobTimer = null;
  let selectedFile = null;

  const qs = new URLSearchParams(window.location.search);
  const urlSID = qs.get("sid") || "";

  const $ = (id) => document.getElementById(id);
  const t = (key) => (I18N[lang] && I18N[lang][key]) || I18N.en[key] || key;

  function applyI18n() {
    document.querySelectorAll("[data-i18n]").forEach((el) => {
      const key = el.getAttribute("data-i18n");
      if (key) el.textContent = t(key);
    });
    $("btn-lang").textContent = lang === "zh" ? "EN" : "中文";
  }

  function showToast(msg, ok = true) {
    const area = $("toast-area");
    const div = document.createElement("div");
    div.className = "toast " + (ok ? "toast-ok" : "toast-err");
    div.textContent = msg;
    area.appendChild(div);
    setTimeout(() => div.remove(), 4500);
  }

  async function api(action, opts = {}) {
    const params = new URLSearchParams(opts.query || {});
    params.set("action", action);
    const headers = Object.assign({}, opts.headers || {});
    if (urlSID) headers["X-QTS-SID"] = urlSID;
    if (opts.json !== undefined) headers["Content-Type"] = "application/json";

    const resp = await fetch("api.cgi?" + params.toString(), {
      method: opts.method || "GET",
      credentials: "same-origin",
      headers,
      body: opts.json !== undefined ? JSON.stringify(opts.json) : opts.body
    });

    let data = {};
    try {
      data = await resp.json();
    } catch (_) {}

    if (resp.status === 401 || resp.status === 403) {
      $("auth-banner").textContent = t("auth_required") + (data.error ? ` (${data.error})` : "");
      $("auth-banner").classList.remove("hidden");
      throw new Error(data.error || "Unauthorized");
    }
    $("auth-banner").classList.add("hidden");

    if (!resp.ok) {
      throw new Error(data.error || `HTTP ${resp.status}`);
    }
    return data;
  }

  function fmtBytes(n) {
    if (!n || n <= 0) return "-";
    const units = ["B", "KB", "MB", "GB", "TB"];
    let i = 0;
    let v = n;
    while (v >= 1024 && i < units.length - 1) {
      v /= 1024;
      i++;
    }
    return `${v.toFixed(i === 0 ? 0 : 1)} ${units[i]}`;
  }

  function fmtDuration(sec) {
    if (sec === undefined || sec === null || sec < 0) return "-";
    const d = Math.floor(sec / 86400);
    const h = Math.floor((sec % 86400) / 3600);
    const m = Math.floor((sec % 3600) / 60);
    const s = sec % 60;
    const parts = [];
    if (d) parts.push(`${d}d`);
    if (h || d) parts.push(`${h}h`);
    if (m || h || d) parts.push(`${m}m`);
    parts.push(`${s}s`);
    return parts.join(" ");
  }

  function switchTab(name) {
    document.querySelectorAll(".tab").forEach((b) => b.classList.toggle("active", b.dataset.tab === name));
    document.querySelectorAll(".tab-panel").forEach((p) => p.classList.toggle("active", p.id === "tab-" + name));
    if (name === "logs") loadLogs();
  }

  function renderStatus(st) {
    latestStatus = st;
    const host = st.host || {};
    $("hdr-host-sub").textContent = `${host.hostname || "QNAP"} · ${host.qts_model || "QTS"} ${host.qts_version || ""} (${host.uname_arch || host.arch || ""})`;

    const pill = $("hdr-status-pill");
    pill.className = "pill";
    if (!st.binary_installed) {
      pill.classList.add("pill-warning");
      pill.textContent = t("status_uninstalled");
    } else if (st.running) {
      pill.classList.add("pill-success");
      pill.textContent = `${t("status_running")} · v${st.binary_version || "?"}`;
    } else if (!st.configured) {
      pill.classList.add("pill-warning");
      pill.textContent = t("status_unconfigured");
    } else {
      pill.classList.add("pill-danger");
      pill.textContent = t("status_stopped");
    }

    $("ov-running").textContent = st.running ? t("status_running") : t("status_stopped");
    $("ov-pid").textContent = st.running && st.pid ? String(st.pid) : "-";
    $("ov-uptime").textContent = st.running ? fmtDuration(st.uptime_sec) : "-";
    $("ov-mem").textContent = st.running ? fmtBytes(st.memory_rss_bytes) : "-";
    $("ov-health").textContent = st.running ? (st.health_ok ? `OK (${st.health_msg || "ok"})` : `Fail (${st.health_msg || "-"})`) : "-";
    $("ov-configured").textContent = st.configured ? t("yes") : t("no");

    $("btn-start").disabled = st.running || !st.binary_installed || !st.configured;
    $("btn-stop").disabled = !st.running;
    $("btn-restart").disabled = !st.binary_installed || !st.configured;

    $("ov-bin-ver").textContent = st.binary_installed ? `v${st.binary_version || "unknown"}` : t("not_installed");
    $("ov-bin-src").textContent = (st.binary_meta && st.binary_meta.source) || "-";
    $("ov-bin-size").textContent = st.binary_installed ? fmtBytes(st.binary_size) : "-";
    $("ov-updated-at").textContent = (st.binary_meta && st.binary_meta.updated_at) || "-";
    $("ov-sha256").textContent = st.binary_sha256 || "-";
    if (st.fingerprint) $("ov-fp").textContent = st.fingerprint;

    $("up-bin-ver").textContent = st.binary_installed ? `v${st.binary_version || "unknown"}` : t("not_installed");
    $("up-bin-src").textContent = (st.binary_meta && st.binary_meta.source) || "-";
    $("up-bak-ver").textContent = st.has_backup ? `v${st.backup_version || "backup"}` : "-";
    $("up-updated-at").textContent = (st.binary_meta && st.binary_meta.updated_at) || "-";
    $("btn-rollback").classList.toggle("hidden", !st.has_backup);

    $("host-model").textContent = `${host.qts_model || "-"} / QTS ${host.qts_version || "-"}`;
    $("host-arch").textContent = `${host.arch || "-"} (${host.uname_arch || "-"}) · glibc=${host.has_glibc ? "yes" : "no"}`;
    $("host-qpkg").textContent = st.qpkg_version || "-";
    $("host-docker").textContent = host.docker_socket || "not found";
    $("host-smartctl").textContent = host.smartctl_path || "not found (Entware opkg smartmontools)";
    $("host-sensors").textContent = (host.hwmon_sensors && host.hwmon_sensors.length)
      ? host.hwmon_sensors.map((s) => `${s.name}(${s.device})`).join(", ")
      : "none";

    renderVolumeChips(host.volumes || []);
    renderChecklist(st);
  }

  function renderVolumeChips(volumes) {
    const hostVolEl = $("host-volumes");
    hostVolEl.innerHTML = "";
    const cfgChips = $("cfg-volume-chips");
    cfgChips.innerHTML = "";
    $("cfg-volume-suggestions").classList.toggle("hidden", volumes.length === 0);

    volumes.forEach((v) => {
      const chip = document.createElement("span");
      chip.className = "chip mono";
      chip.textContent = `${v.path} (${fmtBytes(v.free_bytes)} free / ${fmtBytes(v.total_bytes)})`;
      hostVolEl.appendChild(chip);

      const btn = document.createElement("button");
      btn.type = "button";
      btn.className = "chip mono";
      btn.textContent = v.path;
      btn.title = "Click to set FILESYSTEM (Shift+Click to append to EXTRA_FILESYSTEMS)";
      btn.addEventListener("click", (e) => {
        if (e.shiftKey || ($("cfg-filesystem").value.trim() && $("cfg-filesystem").value.trim() !== v.path)) {
          const cur = $("cfg-extra-fs").value.split(",").map((s) => s.trim()).filter(Boolean);
          if (!$("cfg-filesystem").value.trim()) {
            $("cfg-filesystem").value = v.path;
          } else if (!cur.includes(v.path) && $("cfg-filesystem").value.trim() !== v.path) {
            cur.push(v.path);
            $("cfg-extra-fs").value = cur.join(",");
          }
        } else {
          $("cfg-filesystem").value = v.path;
        }
      });
      cfgChips.appendChild(btn);
    });
  }

  function renderChecklist(st) {
    const show = !st.binary_installed || !st.configured || !st.running;
    $("setup-checklist").classList.toggle("hidden", !show);
    if (!show) return;

    const list = $("checklist-items");
    list.innerHTML = "";
    const steps = [
      { done: st.binary_installed, label: t("step_bin"), tab: "updates" },
      { done: st.configured, label: t("step_cfg"), tab: "config" },
      { done: st.running, label: t("step_run"), action: "start" }
    ];
    steps.forEach((s) => {
      const row = document.createElement("div");
      row.className = "check-row";
      const left = document.createElement("span");
      left.textContent = (s.done ? "[OK] " : "[ ] ") + s.label;
      left.className = s.done ? "muted" : "";
      row.appendChild(left);

      if (!s.done) {
        const btn = document.createElement("button");
        btn.type = "button";
        btn.className = "btn btn-primary btn-sm";
        if (s.tab) {
          btn.textContent = t("btn_go_setup");
          btn.addEventListener("click", () => switchTab(s.tab));
        } else {
          btn.textContent = t("btn_start");
          btn.disabled = !st.binary_installed || !st.configured;
          btn.addEventListener("click", () => runServiceOp("start"));
        }
        row.appendChild(btn);
      }
      list.appendChild(row);
    });
  }

  async function refreshStatus(withFingerprint = false) {
    try {
      const data = await api("status", { query: withFingerprint ? { fp: "1" } : {} });
      renderStatus(data.status);
    } catch (err) {
      showToast(err.message, false);
    }
  }

  function addCustomEnvRow(k = "", v = "") {
    const row = document.createElement("div");
    row.className = "env-row";
    const keyInput = document.createElement("input");
    keyInput.type = "text";
    keyInput.placeholder = "ENV_KEY";
    keyInput.className = "env-key mono";
    keyInput.value = k;

    const valInput = document.createElement("input");
    valInput.type = "text";
    valInput.placeholder = "value";
    valInput.className = "env-val mono";
    valInput.value = v;

    const rm = document.createElement("button");
    rm.type = "button";
    rm.className = "btn btn-danger btn-sm";
    rm.textContent = "×";
    rm.addEventListener("click", () => row.remove());

    row.append(keyInput, valInput, rm);
    $("custom-env-list").appendChild(row);
  }

  function populateConfig(cfg) {
    $("cfg-listen").value = cfg.listen || "";
    $("cfg-key").value = cfg.key || "";
    $("cfg-hub-url").value = cfg.hub_url || "";
    $("cfg-token").value = cfg.token || "";
    $("cfg-disable-ssh").checked = !!cfg.disable_ssh;
    $("cfg-system-name").value = cfg.system_name || "";
    $("cfg-all-proxy").value = cfg.all_proxy || "";
    $("cfg-ca-cert").value = cfg.ca_cert_file || "";
    $("cfg-hub-headers").value = cfg.hub_headers || "";

    $("cfg-filesystem").value = cfg.filesystem || "";
    $("cfg-extra-fs").value = cfg.extra_filesystems || "";
    $("cfg-disk-cache").value = cfg.disk_usage_cache || "";

    $("cfg-docker-host").value = cfg.docker_host || "";
    $("cfg-docker-timeout").value = cfg.docker_timeout || "";
    $("cfg-docker-img-chk").value = cfg.docker_image_check || "";
    $("cfg-exclude-containers").value = cfg.exclude_containers || "";
    $("cfg-container-label").value = cfg.container_name_label || "";

    $("cfg-smart-devices").value = cfg.smart_devices || "";
    $("cfg-exclude-smart").value = cfg.exclude_smart || "";
    $("cfg-smart-interval").value = cfg.smart_interval || "";
    $("cfg-sensors").value = cfg.sensors || "";
    $("cfg-primary-sensor").value = cfg.primary_sensor || "";
    $("cfg-fan-sensors").value = cfg.fan_sensors || "";
    $("cfg-sensors-timeout").value = cfg.sensors_timeout || "";
    $("cfg-skip-gpu").checked = !!cfg.skip_gpu;
    $("cfg-gpu-collector").value = cfg.gpu_collector || "";

    $("cfg-nics").value = cfg.nics || "";
    $("cfg-mem-calc").value = cfg.mem_calc || "";
    $("cfg-log-level").value = cfg.log_level || "info";
    $("cfg-skip-systemd").checked = !!cfg.skip_systemd;
    $("cfg-skip-wifi").checked = !!cfg.skip_wifi;
    $("cfg-auto-start").checked = !!cfg.auto_start;
    $("cfg-extra-args").value = cfg.extra_args || "";

    $("upd-source").value = cfg.download_source || "github";
    $("upd-custom-prefix").value = cfg.custom_mirror_prefix || "";
    $("upd-http-proxy").value = cfg.http_proxy || "";
    $("upd-prefer-glibc").checked = !!cfg.prefer_glibc;
    toggleCustomPrefix();

    $("custom-env-list").innerHTML = "";
    (cfg.custom_env || []).forEach((p) => addCustomEnvRow(p.key, p.value));
  }

  function collectConfig() {
    const customEnv = [];
    $("custom-env-list").querySelectorAll(".env-row").forEach((row) => {
      const k = row.querySelector(".env-key").value.trim();
      const v = row.querySelector(".env-val").value;
      if (k) customEnv.push({ key: k, value: v });
    });

    return {
      auto_start: $("cfg-auto-start").checked,
      download_source: $("upd-source").value,
      custom_mirror_prefix: $("upd-custom-prefix").value.trim(),
      prefer_glibc: $("upd-prefer-glibc").checked,
      http_proxy: $("upd-http-proxy").value.trim(),
      extra_args: $("cfg-extra-args").value.trim(),

      listen: $("cfg-listen").value.trim(),
      key: $("cfg-key").value.trim(),
      hub_url: $("cfg-hub-url").value.trim(),
      token: $("cfg-token").value.trim(),
      disable_ssh: $("cfg-disable-ssh").checked,
      system_name: $("cfg-system-name").value.trim(),
      all_proxy: $("cfg-all-proxy").value.trim(),
      ca_cert_file: $("cfg-ca-cert").value.trim(),
      hub_headers: $("cfg-hub-headers").value.trim(),

      filesystem: $("cfg-filesystem").value.trim(),
      extra_filesystems: $("cfg-extra-fs").value.trim(),
      disk_usage_cache: $("cfg-disk-cache").value.trim(),

      docker_host: $("cfg-docker-host").value.trim(),
      docker_timeout: $("cfg-docker-timeout").value.trim(),
      docker_image_check: $("cfg-docker-img-chk").value,
      exclude_containers: $("cfg-exclude-containers").value.trim(),
      container_name_label: $("cfg-container-label").value.trim(),

      smart_devices: $("cfg-smart-devices").value.trim(),
      exclude_smart: $("cfg-exclude-smart").value.trim(),
      smart_interval: $("cfg-smart-interval").value.trim(),
      sensors: $("cfg-sensors").value.trim(),
      primary_sensor: $("cfg-primary-sensor").value.trim(),
      fan_sensors: $("cfg-fan-sensors").value.trim(),
      sensors_timeout: $("cfg-sensors-timeout").value.trim(),
      skip_gpu: $("cfg-skip-gpu").checked,
      gpu_collector: $("cfg-gpu-collector").value.trim(),

      nics: $("cfg-nics").value.trim(),
      mem_calc: $("cfg-mem-calc").value.trim(),
      log_level: $("cfg-log-level").value,
      skip_systemd: $("cfg-skip-systemd").checked,
      skip_wifi: $("cfg-skip-wifi").checked,
      custom_env: customEnv
    };
  }

  async function loadConfig() {
    try {
      const data = await api("config");
      populateConfig(data.config);
    } catch (err) {
      showToast(err.message, false);
    }
  }

  async function saveConfig(restart) {
    try {
      const data = await api("config", {
        method: "POST",
        json: { config: collectConfig(), restart }
      });
      if (data.status) renderStatus(data.status);
      if (data.restart_error) {
        showToast(data.restart_error, false);
      } else {
        showToast(restart ? "Saved & restarted" : "Configuration saved", true);
      }
    } catch (err) {
      showToast(err.message, false);
    }
  }

  async function runServiceOp(op) {
    try {
      const data = await api("service", { method: "POST", json: { op } });
      if (data.status) renderStatus(data.status);
      showToast(`Service ${op} succeeded`, true);
    } catch (err) {
      showToast(err.message, false);
      refreshStatus();
    }
  }

  async function fetchReleases() {
    const listEl = $("releases-list");
    listEl.textContent = "Loading...";
    try {
      // Persist download source settings quickly if needed by querying directly
      const data = await api("releases", {
        query: {
          source: $("upd-source").value,
          custom_prefix: $("upd-custom-prefix").value.trim(),
          prefer_glibc: $("upd-prefer-glibc").checked ? "1" : "0"
        }
      });
      renderReleases(data.releases || []);
    } catch (err) {
      listEl.textContent = err.message;
      showToast(err.message, false);
    }
  }

  function renderReleases(releases) {
    const listEl = $("releases-list");
    listEl.innerHTML = "";
    if (!releases.length) {
      const empty = document.createElement("div");
      empty.className = "muted small";
      empty.textContent = "No releases found.";
      listEl.appendChild(empty);
      return;
    }

    const installedVer = latestStatus && latestStatus.binary_installed ? latestStatus.binary_version : "";

    releases.forEach((rel, idx) => {
      const item = document.createElement("div");
      item.className = "release-item";

      const left = document.createElement("div");
      const title = document.createElement("div");
      title.className = "release-title";
      title.textContent = rel.tag_name;

      if (idx === 0 && !rel.prerelease) {
        const b = document.createElement("span");
        b.className = "pill pill-info";
        b.textContent = t("latest_tag");
        title.appendChild(b);
      }
      if (rel.prerelease) {
        const b = document.createElement("span");
        b.className = "pill pill-warning";
        b.textContent = t("prerelease");
        title.appendChild(b);
      }
      if (installedVer && installedVer === rel.version) {
        const b = document.createElement("span");
        b.className = "pill pill-success";
        b.textContent = t("current_tag");
        title.appendChild(b);
      }

      const sub = document.createElement("div");
      sub.className = "muted small mono";
      sub.textContent = rel.asset
        ? `${rel.asset.name} · ${fmtBytes(rel.asset.size)} · ${(rel.published_at || "").slice(0, 10)}`
        : "No matching asset for this architecture";

      left.append(title, sub);

      const btn = document.createElement("button");
      btn.type = "button";
      btn.className = "btn btn-primary btn-sm";
      btn.textContent = t("install_ver");
      btn.disabled = !rel.asset;
      btn.addEventListener("click", () => startDownloadJob({ tag: rel.tag_name }));

      item.append(left, btn);
      listEl.appendChild(item);
    });
  }

  async function startDownloadJob(req) {
    try {
      // Save source/proxy preferences before starting background job
      await api("config", { method: "POST", json: { config: collectConfig(), restart: false } });
      const data = await api("download", {
        method: "POST",
        json: Object.assign({
          source: $("upd-source").value,
          custom_prefix: $("upd-custom-prefix").value.trim(),
          prefer_glibc: $("upd-prefer-glibc").checked
        }, req)
      });
      renderJob(data.job);
      pollJob();
    } catch (err) {
      showToast(err.message, false);
    }
  }

  function renderJob(job) {
    const card = $("download-job-card");
    if (!job || job.state === "idle") {
      card.classList.add("hidden");
      return;
    }
    card.classList.remove("hidden");
    $("job-stage").textContent = job.stage || job.state;
    $("job-stage").className = "pill " + (job.state === "error" ? "pill-danger" : job.state === "done" ? "pill-success" : "pill-info");
    $("job-message").textContent = job.error || job.message || "";
    $("job-bar").style.width = (job.percent || 0) + "%";
    $("job-bytes").textContent = job.total_bytes > 0
      ? `${fmtBytes(job.bytes_downloaded)} / ${fmtBytes(job.total_bytes)} (${job.percent || 0}%)`
      : (job.bytes_downloaded > 0 ? fmtBytes(job.bytes_downloaded) : "");
  }

  function pollJob() {
    if (jobTimer) clearInterval(jobTimer);
    jobTimer = setInterval(async () => {
      try {
        const data = await api("download_status");
        renderJob(data.job);
        if (data.job.state === "done" || data.job.state === "error") {
          clearInterval(jobTimer);
          jobTimer = null;
          if (data.status) renderStatus(data.status);
          else refreshStatus();
          showToast(data.job.state === "done" ? (data.job.message || "Update complete") : (data.job.error || "Update failed"), data.job.state === "done");
        }
      } catch (_) {}
    }, 600);
  }

  function uploadFile() {
    if (!selectedFile) return;
    const form = new FormData();
    form.append("file", selectedFile);

    const params = new URLSearchParams({ action: "upload" });
    const xhr = new XMLHttpRequest();
    xhr.open("POST", "api.cgi?" + params.toString(), true);
    xhr.withCredentials = true;
    if (urlSID) xhr.setRequestHeader("X-QTS-SID", urlSID);

    $("upload-bar-wrap").classList.remove("hidden");
    $("btn-upload").disabled = true;

    xhr.upload.onprogress = (e) => {
      if (e.lengthComputable) {
        const pct = Math.round((e.loaded * 100) / e.total);
        $("upload-bar").style.width = pct + "%";
        $("upload-progress-text").textContent = `${pct}% (${fmtBytes(e.loaded)} / ${fmtBytes(e.total)})`;
      }
    };

    xhr.onload = () => {
      $("btn-upload").disabled = false;
      let data = {};
      try { data = JSON.parse(xhr.responseText); } catch (_) {}
      if (xhr.status >= 200 && xhr.status < 300) {
        if (data.status) renderStatus(data.status);
        showToast(data.message || "Uploaded & installed", true);
        $("upload-progress-text").textContent = data.message || "Done";
      } else {
        showToast(data.error || `Upload failed (${xhr.status})`, false);
        $("upload-progress-text").textContent = data.error || "Failed";
      }
    };

    xhr.onerror = () => {
      $("btn-upload").disabled = false;
      showToast("Network error during upload", false);
    };

    xhr.send(form);
  }

  async function loadLogs() {
    try {
      const data = await api("logs", {
        query: { kind: currentLogKind, lines: $("log-lines").value }
      });
      const viewer = $("log-viewer");
      viewer.textContent = (data.lines && data.lines.length) ? data.lines.join("\n") : "(empty)";
      viewer.scrollTop = viewer.scrollHeight;
    } catch (err) {
      showToast(err.message, false);
    }
  }

  function toggleCustomPrefix() {
    $("upd-custom-prefix-wrap").classList.toggle("hidden", $("upd-source").value !== "custom");
  }

  function bindEvents() {
    document.querySelectorAll(".tab").forEach((b) => b.addEventListener("click", () => switchTab(b.dataset.tab)));
    document.querySelectorAll("[data-goto]").forEach((b) => b.addEventListener("click", () => switchTab(b.dataset.goto)));

    $("btn-lang").addEventListener("click", () => {
      lang = lang === "zh" ? "en" : "zh";
      localStorage.setItem("beszel_qnap_lang", lang);
      applyI18n();
      if (latestStatus) renderStatus(latestStatus);
    });

    $("btn-refresh-status").addEventListener("click", () => refreshStatus(true));
    $("btn-start").addEventListener("click", () => runServiceOp("start"));
    $("btn-stop").addEventListener("click", () => runServiceOp("stop"));
    $("btn-restart").addEventListener("click", () => runServiceOp("restart"));

    $("btn-fp-view").addEventListener("click", async () => {
      try {
        const data = await api("fingerprint", { method: "POST", json: { op: "view" } });
        $("ov-fp").textContent = data.fingerprint || "-";
      } catch (err) {
        showToast(err.message, false);
      }
    });

    $("btn-fp-reset").addEventListener("click", async () => {
      if (!confirm(t("confirm_fp_reset"))) return;
      try {
        const data = await api("fingerprint", { method: "POST", json: { op: "reset" } });
        showToast(data.message || "Fingerprint reset", true);
        refreshStatus(true);
      } catch (err) {
        showToast(err.message, false);
      }
    });

    $("config-form").addEventListener("submit", (e) => {
      e.preventDefault();
      saveConfig(false);
    });
    $("btn-cfg-save-restart").addEventListener("click", () => saveConfig(true));
    $("btn-cfg-reload").addEventListener("click", loadConfig);
    $("btn-add-env").addEventListener("click", () => addCustomEnvRow());
    $("btn-fill-docker").addEventListener("click", () => {
      const sock = (latestStatus && latestStatus.host && latestStatus.host.docker_socket) || "/var/run/docker.sock";
      $("cfg-docker-host").value = sock.startsWith("unix://") ? sock : "unix://" + sock;
    });

    $("upd-source").addEventListener("change", toggleCustomPrefix);
    $("btn-fetch-releases").addEventListener("click", fetchReleases);
    $("btn-download-url").addEventListener("click", () => {
      const u = $("upd-custom-url").value.trim();
      if (!u) return;
      startDownloadJob({ custom_url: u });
    });

    $("btn-rollback").addEventListener("click", async () => {
      if (!confirm(t("confirm_rollback"))) return;
      try {
        const data = await api("rollback", { method: "POST" });
        if (data.status) renderStatus(data.status);
        showToast("Rolled back successfully", true);
      } catch (err) {
        showToast(err.message, false);
      }
    });

    const dz = $("dropzone");
    const fileInput = $("file-upload");
    const selectFile = (f) => {
      selectedFile = f || null;
      $("dropzone-label").textContent = f ? `${f.name} (${fmtBytes(f.size)})` : t("drop_hint");
      $("btn-upload").disabled = !f;
    };
    fileInput.addEventListener("change", () => selectFile(fileInput.files[0]));
    dz.addEventListener("dragover", (e) => { e.preventDefault(); dz.classList.add("dragover"); });
    dz.addEventListener("dragleave", () => dz.classList.remove("dragover"));
    dz.addEventListener("drop", (e) => {
      e.preventDefault();
      dz.classList.remove("dragover");
      if (e.dataTransfer.files && e.dataTransfer.files[0]) selectFile(e.dataTransfer.files[0]);
    });
    $("btn-upload").addEventListener("click", uploadFile);

    document.querySelectorAll(".log-kind").forEach((b) => {
      b.addEventListener("click", () => {
        currentLogKind = b.dataset.kind;
        document.querySelectorAll(".log-kind").forEach((x) => {
          x.classList.toggle("active", x === b);
          x.className = "btn btn-sm log-kind " + (x === b ? "btn-secondary active" : "btn-ghost");
        });
        loadLogs();
      });
    });
    $("log-lines").addEventListener("change", loadLogs);
    $("btn-log-refresh").addEventListener("click", loadLogs);
    $("btn-log-clear").addEventListener("click", async () => {
      if (!confirm(t("confirm_clear_log"))) return;
      try {
        await api("logs_clear", { method: "POST", query: { kind: currentLogKind } });
        loadLogs();
      } catch (err) {
        showToast(err.message, false);
      }
    });
    $("log-autorefresh").addEventListener("change", () => {
      if (logTimer) { clearInterval(logTimer); logTimer = null; }
      if ($("log-autorefresh").checked) {
        logTimer = setInterval(loadLogs, 3000);
      }
    });
  }

  async function init() {
    applyI18n();
    bindEvents();
    await Promise.all([refreshStatus(true), loadConfig()]);
    try {
      const data = await api("download_status");
      if (data.job && data.job.state === "running") {
        renderJob(data.job);
        pollJob();
      }
    } catch (_) {}
  }

  init();
})();
