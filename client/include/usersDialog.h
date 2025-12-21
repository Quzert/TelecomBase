#pragma once

#include <QDialog>
#include <QList>

#include "models.h"

class ApiClient;
class QTableWidget;
class QAction;

class UsersDialog final : public QDialog {
    Q_OBJECT

public:
    explicit UsersDialog(ApiClient* apiClient, QWidget* parent = nullptr);

private:
    void buildUi();
    void reload();
    qint64 selectedId() const;

    void approveUser();
    void disableUser();
    void deleteUser();

    ApiClient* apiClient_;
    QList<UserItem> users_;

    QTableWidget* table_;
    QAction* approveAction_;
    QAction* disableAction_;
    QAction* deleteAction_;
    QAction* refreshAction_;
};
