#pragma once

#include <QMainWindow>
#include <QList>
#include <QString>

#include "models.h"

class ApiClient;

class MainWindow final : public QMainWindow {
    Q_OBJECT

public:
    explicit MainWindow(ApiClient* apiClient, const QString& username, const QString& role, QWidget* parent = nullptr);

    void setSession(const QString& username, const QString& role);

signals:
    void logoutRequested();

private:
    void buildUi(const QString& username, const QString& role);
    void loadDevices();
    qint64 selectedDeviceId() const;

    ApiClient* apiClient_;

    QList<DeviceItem> devices_;

    class QLineEdit* searchEdit_;
    class QTableWidget* table_;
    class QAction* addAction_;
    class QAction* editAction_;
    class QAction* vendorsAction_;
    class QAction* modelsAction_;
    class QAction* locationsAction_;
    class QAction* usersAction_;
    class QAction* refreshAction_;
    class QAction* deleteAction_;
    class QAction* logoutAction_;
    class QAction* themeAction_;
};
