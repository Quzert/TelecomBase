#include "apiClient.h"
#include "authDialog.h"
#include "mainWindow.h"

#include <QApplication>

int main(int argc, char* argv[]) {
    QApplication app(argc, argv);
    app.setQuitOnLastWindowClosed(false);

    ApiClient apiClient;

    AuthDialog authDialog(&apiClient);
    if (authDialog.exec() != QDialog::Accepted) {
        return 0;
    }

    apiClient.setToken(authDialog.token());

    MainWindow mainWindow(&apiClient, authDialog.username(), authDialog.role());

    QObject::connect(&mainWindow, &MainWindow::logoutRequested, [&]() {
        const QString prevToken = apiClient.token();
        apiClient.setToken(QString());
        mainWindow.hide();

        AuthDialog dlg(&apiClient);
        if (dlg.exec() != QDialog::Accepted) {
            apiClient.setToken(prevToken);
            mainWindow.show();
            return;
        }

        apiClient.setToken(dlg.token());
        mainWindow.setSession(dlg.username(), dlg.role());
        mainWindow.show();
    });

    mainWindow.show();
    return app.exec();
}
