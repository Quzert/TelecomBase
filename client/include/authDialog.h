#pragma once

#include <QDialog>

class ApiClient;

class AuthDialog final : public QDialog {
    Q_OBJECT

public:
    explicit AuthDialog(ApiClient* apiClient, QWidget* parent = nullptr);

    QString token() const;
    QString username() const;
    QString role() const;

private slots:
    void onLoginClicked();
    void onRegisterClicked();

private:
    void setBusy(bool busy);

    ApiClient* apiClient_;

    QString token_;
    QString username_;
    QString role_;

    class QTabWidget* tabWidget_;

    class QLineEdit* loginUsername_;
    class QLineEdit* loginPassword_;

    class QLineEdit* regUsername_;
    class QLineEdit* regPassword_;
    class QLineEdit* regPassword2_;

    class QPushButton* loginButton_;
    class QPushButton* registerButton_;
    class QPushButton* cancelButton_;
};
