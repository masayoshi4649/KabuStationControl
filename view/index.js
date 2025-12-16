/**
 * 起動パネルの画面初期化を行う関数。
 *
 * 2 つのボタンにクリックイベントを設定し、axios で API を呼び出します。
 * 実行結果は通知（iziToast）で表示します。
 *
 * @function initializeBootPanel
 * @param {void} - 引数はありません。
 * @returns {void} 返り値はありません。
 */
function initializeBootPanel() {
  const bootPanel = document.getElementById("boot-panel");

  const bootAuthKabusButton = document.getElementById("btn-bootauthkabus");
  const bootAppButton = document.getElementById("btn-bootapp");

  if (!bootPanel || !bootAuthKabusButton || !bootAppButton) {
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
      buttons: [bootAuthKabusButton, bootAppButton],
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
 * @param {HTMLButtonElement[]} params.buttons - 有効/無効を切り替えるボタン配列です。
 * @param {string} params.actionLabel - 画面表示用のアクション名です。
 * @param {string} params.path - 呼び出す API パスです。
 * @returns {Promise<void>} 返り値はありません。
 */
async function runAction({ panelElement, buttons, actionLabel, path }) {
  if (panelElement.classList.contains("is-loading")) {
    return;
  }

  setLoadingState(panelElement, buttons, true);

  try {
    const response = await axios.get(path, { timeout: 120000 });
    const data = response?.data ?? {};

    if (data.ok) {
      iziToast.success({
        title: "成功",
        message: data.message || `${actionLabel} が完了しました。`,
        position: "topRight",
      });
      return;
    }

    iziToast.error({
      title: "失敗",
      message: data.message || `${actionLabel} に失敗しました。`,
      position: "topRight",
    });
  } catch (error) {
    const detail = normalizeAxiosError(error);
    const message =
      detail.data?.message || detail.data?.error || detail.message || `${actionLabel} のリクエスト中にエラーが発生しました。`;

    iziToast.error({
      title: "通信エラー",
      message,
      position: "topRight",
    });
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
