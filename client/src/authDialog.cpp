#include "authDialog.h"

#include "apiClient.h"
#include "messageBoxUtils.h"

#include <QFormLayout>
#include <QHBoxLayout>
#include <QLineEdit>
#include <QMessageBox>
#include <QPushButton>
#include <QTabWidget>
#include <QVBoxLayout>

AuthDialog::AuthDialog(ApiClient* apiClient, QWidget* parent)
    : QDialog(parent)
    , apiClient_(apiClient) {
    setWindowTitle("TelecomBase — вход");
    setModal(true);

    tabWidget_ = new QTabWidget(this);

    QWidget* loginTab = new QWidget(this);
    QFormLayout* loginForm = new QFormLayout(loginTab);
    loginUsername_ = new QLineEdit(loginTab);
    loginPassword_ = new QLineEdit(loginTab);
    loginPassword_->setEchoMode(QLineEdit::Password);
    loginForm->addRow("Логин", loginUsername_);
    loginForm->addRow("Пароль", loginPassword_);
    tabWidget_->addTab(loginTab, "Вход");

    QWidget* registerTab = new QWidget(this);
    QFormLayout* regForm = new QFormLayout(registerTab);
    regUsername_ = new QLineEdit(registerTab);
    regPassword_ = new QLineEdit(registerTab);
    regPassword_->setEchoMode(QLineEdit::Password);
    regPassword2_ = new QLineEdit(registerTab);
    regPassword2_->setEchoMode(QLineEdit::Password);
    regForm->addRow("Логин", regUsername_);
    regForm->addRow("Пароль", regPassword_);
    regForm->addRow("Повтор пароля", regPassword2_);
    tabWidget_->addTab(registerTab, "Регистрация");

    loginButton_ = new QPushButton("Войти", this);
    registerButton_ = new QPushButton("Зарегистрироваться", this);
    cancelButton_ = new QPushButton("Отмена", this);

    QHBoxLayout* buttons = new QHBoxLayout();
    buttons->addStretch(1);
    buttons->addWidget(loginButton_);
    buttons->addWidget(registerButton_);
    buttons->addWidget(cancelButton_);

    QVBoxLayout* root = new QVBoxLayout(this);
    root->setContentsMargins(16, 16, 16, 16);
    root->setSpacing(12);
    root->addWidget(tabWidget_);
    root->addLayout(buttons);

    connect(loginButton_, &QPushButton::clicked, this, &AuthDialog::onLoginClicked);
    connect(registerButton_, &QPushButton::clicked, this, &AuthDialog::onRegisterClicked);
    connect(cancelButton_, &QPushButton::clicked, this, &AuthDialog::reject);

    auto updateButtonsForTab = [this]() {
        const int idx = tabWidget_->currentIndex();
        const bool isLogin = (idx == 0);
        loginButton_->setVisible(isLogin);
        registerButton_->setVisible(!isLogin);
        if (isLogin) {
            loginButton_->setDefault(true);
            registerButton_->setDefault(false);
        } else {
            registerButton_->setDefault(true);
            loginButton_->setDefault(false);
        }
    };

    connect(tabWidget_, &QTabWidget::currentChanged, this, [updateButtonsForTab](int) { updateButtonsForTab(); });
    updateButtonsForTab();

    setFixedWidth(460);
}

QString AuthDialog::token() const {
    return token_;
}

QString AuthDialog::username() const {
    return username_;
}

QString AuthDialog::role() const {
    return role_;
}

void AuthDialog::setBusy(bool busy) {
    tabWidget_->setEnabled(!busy);
    loginButton_->setEnabled(!busy);
    registerButton_->setEnabled(!busy);
    cancelButton_->setEnabled(!busy);
}

void AuthDialog::onLoginClicked() {
    if (!apiClient_) {
        UiUtils::critical(this, "Ошибка", "API клиент не инициализирован");
        return;
    }

    setBusy(true);
    const auto res = apiClient_->login(loginUsername_->text().trimmed(), loginPassword_->text());
    setBusy(false);

    if (!res.ok) {
        QString msg = res.error;
        if (msg == "account_pending_approval") {
            msg = "Аккаунт ожидает подтверждения администратором";
        }
        UiUtils::warning(this, "Не удалось войти", msg.isEmpty() ? "Ошибка" : msg);
        return;
    }

    token_ = res.token;
    username_ = res.username;
    role_ = res.role;
    accept();
}

void AuthDialog::onRegisterClicked() {
    if (!apiClient_) {
        UiUtils::critical(this, "Ошибка", "API клиент не инициализирован");
        return;
    }

    const QString p1 = regPassword_->text();
    const QString p2 = regPassword2_->text();
    if (p1 != p2) {
        UiUtils::warning(this, "Ошибка", "Пароли не совпадают");
        return;
    }

    setBusy(true);
    const auto res = apiClient_->registerUser(regUsername_->text().trimmed(), p1);
    setBusy(false);

    if (!res.ok) {
        QString msg = res.error;
        if (msg == "account_pending_approval") {
            msg = "Аккаунт создан, но ожидает подтверждения администратором";
        }
        UiUtils::warning(this, "Не удалось зарегистрироваться", msg.isEmpty() ? "Ошибка" : msg);
        return;
    }

    token_ = res.token;
    username_ = res.username;
    role_ = res.role;
    accept();
}
