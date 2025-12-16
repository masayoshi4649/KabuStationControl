/**
 * 起動パネルの画面初期化を行う関数。
 *
 * 3 つのボタンにクリックイベントを設定し、axios で API を呼び出します。
 * 実行結果はログへ追記し、通知は iziToast で表示します。
 *
 * @function initializeBootPanel
 * @param {void} - 引数はありません。
 * @returns {void} 返り値はありません。
 */
function initializeBootPanel() {
  const bootPanel = document.getElementById("boot-panel");
  const bootLog = document.getElementById("boot-log");

  const bootAuthKabusButton = document.getElementById("btn-bootauthkabus");
  const bootAppButton = document.getElementById("btn-bootapp");

  if (!bootPanel || !bootLog || !bootAuthKabusButton || !bootAppButton) {
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
      logElement: bootLog,
      buttons: [bootAuthKabusButton, bootAppButton],
      actionLabel: "KabuStation 起動認証",
      path: "/bootauthkabus",
    });
  });

  bootAppButton.addEventListener("click", async (event) => {
    event.preventDefault();
    event.stopPropagation();
    await runAction({
      panelElement: bootPanel,
      logElement: bootLog,
      buttons: [bootAuthKabusButton, bootAppButton],
      actionLabel: "TradeWebApp 起動",
      path: "/bootapp",
    });
  });
}

// ----------------------------------------

/**
 * API 呼び出しを実行し、ログと通知を更新する関数。
 *
 * ボタンを無効化し、GET リクエスト完了後に復帰します。
 * 成功時は `ok: true` の JSON を期待し、失敗時はエラー内容を表示します。
 *
 * @function runAction
 * @param {Object} params - 実行に必要なパラメータ群です。
 * @param {HTMLElement} params.panelElement - ローディング状態を付与する要素です。
 * @param {HTMLElement} params.logElement - 実行ログを追記する要素です。
 * @param {HTMLButtonElement[]} params.buttons - 有効/無効を切り替えるボタン配列です。
 * @param {string} params.actionLabel - 画面表示用のアクション名です。
 * @param {string} params.path - 呼び出す API パスです。
 * @returns {Promise<void>} 返り値はありません。
 */
async function runAction({ panelElement, logElement, buttons, actionLabel, path }) {
  if (panelElement.classList.contains("is-loading")) {
    return;
  }

  setLoadingState(panelElement, buttons, true);
  appendLog(logElement, `${actionLabel} を開始しました。`);
  appendLog(logElement, `GET ${path} を送信します。`);

  try {
    const response = await axios.get(path, { timeout: 120000 });
    const data = response?.data ?? {};

    if (data.ok) {
      iziToast.success({
        title: "成功",
        message: `${actionLabel} が完了しました。`,
        position: "topRight",
      });
      appendLog(logElement, `${actionLabel} が完了しました。`);
      appendLog(logElement, formatJsonForLog(data));
      return;
    }

    iziToast.error({
      title: "失敗",
      message: `${actionLabel} に失敗しました。`,
      position: "topRight",
    });
    appendLog(logElement, `${actionLabel} に失敗しました。`);
    appendLog(logElement, formatJsonForLog(data));
  } catch (error) {
    iziToast.error({
      title: "通信エラー",
      message: `${actionLabel} のリクエスト中にエラーが発生しました。`,
      position: "topRight",
    });

    const detail = normalizeAxiosError(error);
    appendLog(logElement, `${actionLabel} のリクエスト中にエラーが発生しました。`);
    appendLog(logElement, formatJsonForLog(detail));
  } finally {
    setLoadingState(panelElement, buttons, false);
  }
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
  } else {
    panelElement.classList.remove("is-loading");
  }

  buttons.forEach((button) => {
    button.disabled = isLoading;
  });
}

// ----------------------------------------

/**
 * 実行ログへメッセージを追記する関数。
 *
 * @function appendLog
 * @param {HTMLElement} logElement - ログ表示要素です。
 * @param {string} message - 追記するメッセージです。
 * @returns {void} 返り値はありません。
 */
function appendLog(logElement, message) {
  const now = new Date();
  const timestamp = now.toLocaleTimeString("ja-JP", { hour12: false });

  const line = document.createElement("div");
  line.textContent = `[${timestamp}] ${message}`;

  if (logElement.children.length === 1 && logElement.firstElementChild?.dataset?.placeholder === "true") {
    logElement.innerHTML = "";
  }

  logElement.appendChild(line);
  logElement.scrollTop = logElement.scrollHeight;
}

// ----------------------------------------

/**
 * JSON をログ用の文字列へ整形する関数。
 *
 * @function formatJsonForLog
 * @param {Object} data - 整形対象のデータです。
 * @returns {string} 返り値はログ表示向けの 1 行文字列です。
 */
function formatJsonForLog(data) {
  try {
    return JSON.stringify(data);
  } catch {
    return String(data);
  }
}

// ----------------------------------------

/**
 * axios のエラー形式をログ表示用へ正規化する関数。
 *
 * @function normalizeAxiosError
 * @param {unknown} error - axios が投げる例外オブジェクトです。
 * @returns {Object} 返り値はログ出力用の情報オブジェクトです。
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
