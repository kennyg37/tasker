#ifndef DISPLAY_H
#define DISPLAY_H

#include "Task.h"
#include "TodoManager.h"
#include <vector>
#include <string>

class Display {
private:
    // ANSI color codes
    static const std::string COLOR_RESET;
    static const std::string COLOR_GREEN;
    static const std::string COLOR_YELLOW;
    static const std::string COLOR_RED;
    static const std::string COLOR_BLUE;
    static const std::string COLOR_CYAN;
    static const std::string COLOR_BOLD;

    std::string getColorForStatus(TaskStatus status) const;
    void printHeader() const;
    void printSeparator() const;

public:
    Display() = default;

    // Display functions
    void showTasks(const std::vector<Task>& tasks, const std::string& title = "Tasks") const;
    void showStatistics(const TodoManager& manager) const;
    void showHelp() const;
};

#endif // DISPLAY_H
