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
  const requestStatus = document.getElementById("request-status");

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
    await runAction({
      panelElement: bootPanel,
      statusElement: requestStatus,
      buttons: [bootAuthKabusButton, apiAuthButton, bootAppButton],
      clickedButton: bootAuthKabusButton,
      actionLabel: "KabuStation 起動認証",
      path: "/bootauthkabus",
    });
  });

  apiAuthButton.addEventListener("click", async (event) => {
    event.preventDefault();
    event.stopPropagation();
    await runAction({
      panelElement: bootPanel,
      statusElement: requestStatus,
      buttons: [bootAuthKabusButton, apiAuthButton, bootAppButton],
      clickedButton: apiAuthButton,
      actionLabel: "API認証",
      path: "/apiauth",
    });
  });

  bootAppButton.addEventListener("click", async (event) => {
    event.preventDefault();
    event.stopPropagation();
    await runAction({
      panelElement: bootPanel,
      statusElement: requestStatus,
      buttons: [bootAuthKabusButton, apiAuthButton, bootAppButton],
      clickedButton: bootAppButton,
      actionLabel: "TradeWebApp 起動",
      path: "/bootapp",
    });
  });
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
 * @param {HTMLElement|null} params.statusElement - 状態表示要素です（未配置の場合は null です）。
 * @param {HTMLButtonElement[]} params.buttons - 有効/無効を切り替えるボタン配列です。
 * @param {HTMLButtonElement} params.clickedButton - 押下されたボタンです。
 * @param {string} params.actionLabel - 画面表示用のアクション名です。
 * @param {string} params.path - 呼び出す API パスです。
 * @returns {Promise<void>} 返り値はありません。
 */
async function runAction({ panelElement, statusElement, buttons, clickedButton, actionLabel, path }) {
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
  setRequestStatus(statusElement, "loading", `送信中: GET ${path}`);
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
      iziToast.success({
        title: "成功",
        message: data.message || `${actionLabel} が完了しました。`,
        position: "topRight",
      });
      setRequestStatus(statusElement, "success", `完了: ${actionLabel}`);
      return;
    }

    iziToast.error({
      title: "失敗",
      message: data.message || `${actionLabel} に失敗しました。`,
      position: "topRight",
    });
    setRequestStatus(statusElement, "error", `失敗: ${actionLabel}`);
  } catch (error) {
    const detail = normalizeAxiosError(error);
    const message =
      detail.data?.message || detail.data?.error || detail.message || `${actionLabel} のリクエスト中にエラーが発生しました。`;

    iziToast.error({
      title: "通信エラー",
      message,
      position: "topRight",
    });
    setRequestStatus(statusElement, "error", `通信エラー: ${actionLabel}`);
  } finally {
    setLoadingState(panelElement, buttons, false);
    endButtonLoading(clickedButton, originalText);
  }
}

// ----------------------------------------

/**
 * 状態表示を更新する関数。
 *
 * @function setRequestStatus
 * @param {HTMLElement|null} statusElement - 状態表示要素です。
 * @param {"idle"|"loading"|"success"|"error"} status - 状態です。
 * @param {string} message - 表示メッセージです。
 * @returns {void} 返り値はありません。
 */
function setRequestStatus(statusElement, status, message) {
  if (!statusElement) {
    return;
  }

  statusElement.textContent = message;
  statusElement.dataset.status = status;
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
});

if (document.readyState !== "loading") {
  initializeBootPanel();
}
