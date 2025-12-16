param(
    # kabuステーション本体のパス（必要に応じて変えてOK）
    [string]$ExePath = "C:\Users\masay\AppData\Local\kabuStation\KabuS.exe",

    # 待ち時間（秒）
    [int]$TimeoutSeconds = 60
)

Add-Type -AssemblyName UIAutomationClient, UIAutomationTypes

function Start-KabuSIfNeeded {
    param([string]$ExePath)

    $proc = Get-Process -Name "KabuS" -ErrorAction SilentlyContinue
    if (-not $proc) {
        Write-Host "KabuS.exe を起動します: $ExePath"
        Start-Process -FilePath $ExePath | Out-Null
    }
}

function Get-KabuSMainWindow {
    param([int]$TimeoutSeconds = 30)

    $deadline = (Get-Date).AddSeconds($TimeoutSeconds)
    while ((Get-Date) -lt $deadline) {
        $proc = Get-Process -Name "KabuS" -ErrorAction SilentlyContinue
        if ($proc -and $proc.MainWindowHandle -ne 0) {
            $win = [System.Windows.Automation.AutomationElement]::FromHandle($proc.MainWindowHandle)
            return $win
        }
        Start-Sleep -Milliseconds 500
    }
    throw "タイムアウト: KabuS のメインウィンドウが取得できませんでした。"
}

function Wait-And-FindLoginButton {
    param(
        [System.Windows.Automation.AutomationElement]$WindowElement,
        [int]$TimeoutSeconds = 60
    )

    $ctrlTypeProp = [System.Windows.Automation.AutomationElement]::ControlTypeProperty
    $nameProp = [System.Windows.Automation.AutomationElement]::NameProperty
    $frameworkProp = [System.Windows.Automation.AutomationElement]::FrameworkIdProperty

    $condButton = New-Object System.Windows.Automation.PropertyCondition(
        $ctrlTypeProp,
        [System.Windows.Automation.ControlType]::Button
    )
    $condName = New-Object System.Windows.Automation.PropertyCondition(
        $nameProp,
        "ログイン"          # Inspect の Name
    )
    $condFramework = New-Object System.Windows.Automation.PropertyCondition(
        $frameworkProp,
        "Chrome"            # Inspect の FrameworkId
    )

    $andCond = New-Object System.Windows.Automation.AndCondition(
        @($condButton, $condName, $condFramework)
    )

    $deadline = (Get-Date).AddSeconds($TimeoutSeconds)

    while ((Get-Date) -lt $deadline) {

        Write-Host "ログインボタンを探索しています..."

        $btn = $WindowElement.FindFirst(
            [System.Windows.Automation.TreeScope]::Descendants,
            $andCond
        )

        if ($btn) {
            return $btn
        }

        Start-Sleep -Milliseconds 500
    }

    throw "タイムアウト: ログインボタン (Name='ログイン', Button, FrameworkId='Chrome') が見つかりませんでした。"
}

function Invoke-Element {
    param([System.Windows.Automation.AutomationElement]$Element)

    $invokePattern = $Element.GetCurrentPattern(
        [System.Windows.Automation.InvokePattern]::Pattern
    )
    $invokePattern.Invoke()
}

# ---------------- メイン処理 ----------------

Start-KabuSIfNeeded -ExePath $ExePath

Write-Host "KabuS のメインウィンドウを待っています..."
$mainWin = Get-KabuSMainWindow -TimeoutSeconds 30
Write-Host "メインウィンドウ取得:", $mainWin.Current.Name

$loginButton = Wait-And-FindLoginButton -WindowElement $mainWin -TimeoutSeconds $TimeoutSeconds

Write-Host "ログインボタン発見。Invoke 実行..."
Invoke-Element -Element $loginButton

Write-Host "ログインボタンを Invoke しました。"
