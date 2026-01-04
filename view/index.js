/**
 * 起動パネルの画面初期化を行う関数。
 *
 * 3 つのボタンにクリックイベントを設定し、axios で API を呼び出します。
 * 実行結果は通知（iziToast）で表示します。
 *
 * @function initializeBootPanel
 * @param {void} - 引数はありません。
 * @returns {void} 返り値はありません。
 */
function initializeBootPanel() {
  const bootPanel = document.getElementById("boot-panel");

  const bootAuthKabusButton = document.getElementById("btn-bootauthkabus");
  const apiAuthButton = document.getElementById("btn-apiauth");
  const bootAppButton = document.getElementById("btn-bootapp");

  if (!bootPanel || !bootAuthKabusButton || !apiAuthButton || !bootAppButton) {
    return;
  }

  if (bootPanel.dataset.initialized === "true") {
    return;
  }
  bootPanel.dataset.initialized = "true";

  bootAuthKabusButton.addEventListener("click", async (event) => {
    event.preventDefault();
    event.stopPropagation();
    await runActionAndReloadPID({
      panelElement: bootPanel,
      buttons: [bootAuthKabusButton, apiAuthButton, bootAppButton],
      clickedButton: bootAuthKabusButton,
      actionLabel: "KabuStation 起動認証",
      path: "/bootauthkabus",
    });
  });

  apiAuthButton.addEventListener("click", async (event) => {
    event.preventDefault();
    event.stopPropagation();
    await runActionAndReloadPID({
      panelElement: bootPanel,
      buttons: [bootAuthKabusButton, apiAuthButton, bootAppButton],
      clickedButton: apiAuthButton,
      actionLabel: "API認証",
      path: "/apiauth",
    });
  });

  bootAppButton.addEventListener("click", async (event) => {
    event.preventDefault();
    event.stopPropagation();
    await runActionAndReloadPID({
      panelElement: bootPanel,
      buttons: [bootAuthKabusButton, apiAuthButton, bootAppButton],
      clickedButton: bootAppButton,
      actionLabel: "TradeWebApp 起動",
      path: "/bootapp",
    });
  });
}

// ----------------------------------------

/**
 * 強制終了パネルの画面初期化を行う関数。
 *
 * PID 再取得ボタンのクリックイベントと、初回読込時の PID 取得処理を設定します。
 * 取得結果に応じて KabuS と TradeWebApp の強制終了ボタンを有効化します。
 *
 * @function initializeForceKillPanel
 * @param {void} - 引数はありません。
 * @returns {void} 返り値はありません。
 */
function initializeForceKillPanel() {
  const forceKillPanel = document.getElementById("force-kill-panel");
  const killKabusButton = document.getElementById("btn-killkabus");
  const killTradeAppButton = document.getElementById("btn-killtradeapp");
  const reloadPidButton = document.getElementById("btn-reloadpid");

  if (!forceKillPanel || !killKabusButton || !killTradeAppButton || !reloadPidButton) {
    return;
  }

  if (forceKillPanel.dataset.initialized === "true") {
    return;
  }
  forceKillPanel.dataset.initialized = "true";

  updateKillButtonsState({
    killKabusButton,
    killTradeAppButton,
    pidKabus: 0,
    pidTradeApp: 0,
  });

  killKabusButton.addEventListener("click", async (event) => {
    event.preventDefault();
    event.stopPropagation();
    await runActionAndReloadPID({
      panelElement: forceKillPanel,
      buttons: [killKabusButton, killTradeAppButton, reloadPidButton],
      clickedButton: killKabusButton,
      actionLabel: "KabuS 強制終了",
      path: "/killkabus",
    });
  });

  killTradeAppButton.addEventListener("click", async (event) => {
    event.preventDefault();
    event.stopPropagation();
    await runActionAndReloadPID({
      panelElement: forceKillPanel,
      buttons: [killKabusButton, killTradeAppButton, reloadPidButton],
      clickedButton: killTradeAppButton,
      actionLabel: "TradeWebApp 強制終了",
      path: "/killapp",
    });
  });

  reloadPidButton.addEventListener("click", async (event) => {
    event.preventDefault();
    event.stopPropagation();
    await fetchPIDAndUpdateKillButtons({
      panelElement: forceKillPanel,
      killKabusButton,
      killTradeAppButton,
      reloadPidButton,
    });
  });

  fetchPIDAndUpdateKillButtons({
    panelElement: forceKillPanel,
    killKabusButton,
    killTradeAppButton,
    reloadPidButton,
  });
}

// ----------------------------------------

/**
 * 起動パネルの操作後に PID 再取得を行う関数。
 *
 * 起動系リクエストが完了した後、強制終了パネルの PID を再取得します。
 *
 * @function runActionAndReloadPID
 * @param {Object} params - runAction に渡すパラメータ群です。
 * @param {HTMLElement} params.panelElement - ローディング状態を付与する要素です。
 * @param {HTMLButtonElement[]} params.buttons - 有効/無効を切り替えるボタン配列です。
 * @param {HTMLButtonElement} params.clickedButton - 押下されたボタンです。
 * @param {string} params.actionLabel - 画面表示用のアクション名です。
 * @param {string} params.path - 呼び出す API パスです。
 * @returns {Promise<void>} 返り値はありません。
 */
async function runActionAndReloadPID(params) {
  await runAction(params);
  await reloadPIDForForceKillPanel();
}

// ----------------------------------------

/**
 * 強制終了パネルの PID を再取得する関数。
 *
 * 対象要素が見つかった場合のみ fetchPIDAndUpdateKillButtons を実行します。
 *
 * @function reloadPIDForForceKillPanel
 * @param {void} - 引数はありません。
 * @returns {Promise<void>} 返り値はありません。
 */
async function reloadPIDForForceKillPanel() {
  const forceKillPanel = document.getElementById("force-kill-panel");
  const killKabusButton = document.getElementById("btn-killkabus");
  const killTradeAppButton = document.getElementById("btn-killtradeapp");
  const reloadPidButton = document.getElementById("btn-reloadpid");

  if (!forceKillPanel || !killKabusButton || !killTradeAppButton || !reloadPidButton) {
    return;
  }

  await fetchPIDAndUpdateKillButtons({
    panelElement: forceKillPanel,
    killKabusButton,
    killTradeAppButton,
    reloadPidButton,
  });
}

// ----------------------------------------

/**
 * PID を取得し、強制終了ボタンの状態を更新する関数。
 *
 * /pid を axios で呼び出し、取得した PID に応じてボタンを有効化します。
 * 通信失敗時はボタンを無効化したままにします。
 *
 * @function fetchPIDAndUpdateKillButtons
 * @param {Object} params - 実行に必要なパラメータ群です。
 * @param {HTMLElement} params.panelElement - ローディング状態を付与する要素です。
 * @param {HTMLButtonElement} params.killKabusButton - KabuS 強制終了ボタンです。
 * @param {HTMLButtonElement} params.killTradeAppButton - TradeWebApp 強制終了ボタンです。
 * @param {HTMLButtonElement} params.reloadPidButton - PID 再取得ボタンです。
 * @returns {Promise<void>} 返り値はありません。
 */
async function fetchPIDAndUpdateKillButtons({
  panelElement,
  killKabusButton,
  killTradeAppButton,
  reloadPidButton,
}) {
  if (panelElement.classList.contains("is-loading")) {
    return;
  }

  const originalText = beginButtonLoading(reloadPidButton);
  setLoadingState(panelElement, [killKabusButton, killTradeAppButton, reloadPidButton], true);

  let pidKabusValue = 0;
  let pidTradeAppValue = 0;

  try {
    const response = await axios.get("/pid", { timeout: 120000 });
    const data = response?.data ?? {};
    const rawKabus = Number(data.pidKabus);
    const rawTradeApp = Number(data.pidTradeApp);

    if (data.ok && Number.isFinite(rawKabus) && rawKabus > 0) {
      pidKabusValue = rawKabus;
    }
    if (data.ok && Number.isFinite(rawTradeApp) && rawTradeApp > 0) {
      pidTradeAppValue = rawTradeApp;
    }
  } catch (error) {
    const detail = normalizeAxiosError(error);
    const message = detail.data?.message || detail.data?.error || detail.message || "PID の取得に失敗しました。";

    iziToast.error({
      title: "通信エラー",
      message,
      position: "topRight",
    });
  } finally {
    setLoadingState(panelElement, [killKabusButton, killTradeAppButton, reloadPidButton], false);
    endButtonLoading(reloadPidButton, originalText);
    updateKillButtonsState({
      killKabusButton,
      killTradeAppButton,
      pidKabus: pidKabusValue,
      pidTradeApp: pidTradeAppValue,
    });
  }
}

// ----------------------------------------

/**
 * PID の値に応じて強制終了ボタンを有効化する関数。
 *
 * 0 以上の PID が存在する場合のみ、対応するボタンを有効化します。
 *
 * @function updateKillButtonsState
 * @param {Object} params - 更新に必要なパラメータ群です。
 * @param {HTMLButtonElement} params.killKabusButton - KabuS 強制終了ボタンです。
 * @param {HTMLButtonElement} params.killTradeAppButton - TradeWebApp 強制終了ボタンです。
 * @param {number} params.pidKabus - KabuS の PID です。
 * @param {number} params.pidTradeApp - TradeWebApp の PID です。
 * @returns {void} 返り値はありません。
 */
function updateKillButtonsState({ killKabusButton, killTradeAppButton, pidKabus, pidTradeApp }) {
  const isKabusEnabled = Number.isFinite(pidKabus) && pidKabus > 0;
  const isTradeAppEnabled = Number.isFinite(pidTradeApp) && pidTradeApp > 0;

  killKabusButton.disabled = !isKabusEnabled;
  killTradeAppButton.disabled = !isTradeAppEnabled;
}

// ----------------------------------------

/**
 * API 呼び出しを実行し、通知を更新する関数。
 *
 * ボタンを無効化し、GET リクエスト完了後に復帰します。
 * 成功時は `ok: true` の JSON を期待し、失敗時はエラー内容を表示します。
 *
 * @function runAction
 * @param {Object} params - 実行に必要なパラメータ群です。
 * @param {HTMLElement} params.panelElement - ローディング状態を付与する要素です。
 * @param {HTMLButtonElement[]} params.buttons - 有効/無効を切り替えるボタン配列です。
 * @param {HTMLButtonElement} params.clickedButton - 押下されたボタンです。
 * @param {string} params.actionLabel - 画面表示用のアクション名です。
 * @param {string} params.path - 呼び出す API パスです。
 * @returns {Promise<void>} 返り値はありません。
 */
async function runAction({ panelElement, buttons, clickedButton, actionLabel, path }) {
  if (panelElement.classList.contains("is-loading")) {
    iziToast.info({
      title: "実行中",
      message: "処理中のため、完了後に実行してください。",
      position: "topRight",
    });
    return;
  }

  const originalText = beginButtonLoading(clickedButton);
  setLoadingState(panelElement, buttons, true);
  iziToast.info({
    title: "送信",
    message: `GET ${path} を送信しました。`,
    position: "topRight",
    timeout: 1200,
  });

  try {
    const response = await axios.get(path, { timeout: 120000 });
    const data = response?.data ?? {};

    if (data.ok) {
      const message = appendPIDToMessage(data.message || `${actionLabel} が完了しました。`, data);
      iziToast.success({
        title: "成功",
        message,
        position: "topRight",
      });
      return;
    }

    const message = appendPIDToMessage(data.message || `${actionLabel} に失敗しました。`, data);
    iziToast.error({
      title: "失敗",
      message,
      position: "topRight",
    });
  } catch (error) {
    const detail = normalizeAxiosError(error);
    const baseMessage =
      detail.data?.message || detail.data?.error || detail.message || `${actionLabel} のリクエスト中にエラーが発生しました。`;
    const message = appendPIDToMessage(baseMessage, detail.data);

    iziToast.error({
      title: "通信エラー",
      message,
      position: "topRight",
    });
  } finally {
    setLoadingState(panelElement, buttons, false);
    endButtonLoading(clickedButton, originalText);
  }
}

// ----------------------------------------

/**
 * PID 情報をメッセージへ付与する関数。
 *
 * API レスポンスの `pid` と `pids` を見て、存在する場合に `(PID: 1234)` の形式で付与します。
 *
 * @function appendPIDToMessage
 * @param {string} baseMessage - 元となるメッセージです。
 * @param {Object|undefined|null} data - API レスポンスの data オブジェクトです。
 * @returns {string} 返り値は PID 情報を付与したメッセージです。
 */
function appendPIDToMessage(baseMessage, data) {
  const safeMessage = baseMessage || "";
  const source = data || {};
  const pidList = [];

  if (Number.isFinite(source.pid) && source.pid > 0) {
    pidList.push(source.pid);
  }

  if (Array.isArray(source.pids)) {
    source.pids.forEach((pid) => {
      if (!Number.isFinite(pid) || pid <= 0) {
        return;
      }
      if (pidList.includes(pid)) {
        return;
      }
      pidList.push(pid);
    });
  }

  if (pidList.length === 0) {
    return safeMessage;
  }

  return `${safeMessage} (PID: ${pidList.join(", ")})`;
}

// ----------------------------------------

/**
 * ボタンをローディング表示へ切り替える関数。
 *
 * @function beginButtonLoading
 * @param {HTMLButtonElement} button - 対象ボタンです。
 * @returns {string} 返り値は元のボタン表示テキストです。
 */
function beginButtonLoading(button) {
  const originalText = button.textContent || "";
  button.dataset.originalText = originalText;
  button.textContent = `送信中... ${originalText.trim()}`.trim();
  return originalText;
}

// ----------------------------------------

/**
 * ボタンのローディング表示を解除する関数。
 *
 * @function endButtonLoading
 * @param {HTMLButtonElement} button - 対象ボタンです。
 * @param {string} originalText - 元のボタン表示テキストです。
 * @returns {void} 返り値はありません。
 */
function endButtonLoading(button, originalText) {
  button.textContent = originalText || button.dataset.originalText || "";
}

// ----------------------------------------

/**
 * ローディング状態の付与とボタンの無効化を行う関数。
 *
 * @function setLoadingState
 * @param {HTMLElement} panelElement - 状態クラスを付与する要素です。
 * @param {HTMLButtonElement[]} buttons - 有効/無効を切り替えるボタン配列です。
 * @param {boolean} isLoading - ローディング中かどうかです。
 * @returns {void} 返り値はありません。
 */
function setLoadingState(panelElement, buttons, isLoading) {
  if (isLoading) {
    panelElement.classList.add("is-loading");
    panelElement.setAttribute("aria-busy", "true");
  } else {
    panelElement.classList.remove("is-loading");
    panelElement.removeAttribute("aria-busy");
  }

  buttons.forEach((button) => {
    button.disabled = isLoading;
  });
}

// ----------------------------------------

/**
 * axios のエラー形式を表示用へ正規化する関数。
 *
 * @function normalizeAxiosError
 * @param {unknown} error - axios が投げる例外オブジェクトです。
 * @returns {Object} 返り値は表示用の情報オブジェクトです。
 */
function normalizeAxiosError(error) {
  const axiosError = error || {};
  const response = axiosError.response || {};

  return {
    message: axiosError.message || "不明なエラーです。",
    status: response.status,
    data: response.data,
  };
}

// ----------------------------------------

document.addEventListener("DOMContentLoaded", () => {
  initializeBootPanel();
  initializeForceKillPanel();
});

if (document.readyState !== "loading") {
  initializeBootPanel();
  initializeForceKillPanel();
}
