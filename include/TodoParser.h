#ifndef TODOPARSER_H
#define TODOPARSER_H

#include "Task.h"
#include <vector>
#include <string>

class TodoParser {
private:
    std::string filePath;
    
    bool isDateLine(const std::string& line) const;
    TaskStatus parseStatus(const std::string& statusStr) const;
    void parseLine(const std::string& line, const std::string& currentDate, std::vector<Task>& tasks) const;

public:
    TodoParser();

    std::vector<Task> parse();
    
    std::string getFilePathFromEnv() const;
};

#endif // TODOPARSER_H
