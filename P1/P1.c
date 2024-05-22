cat > P1.c <<EOF
#include <stdio.h>
#include <stdlib.h>

extern char **__environ;

int main (int argc, char **argv) {
    FILE *fp = fopen("/volumes/v1/file.txt", "w");
    fprintf(fp, "argv:");
    for (int i = 0; i < argc; i++) {
        fprintf(fp, " %s", argv[i]);
    }
    fprintf(fp, "\n");

    char** envp = __environ;
    fprintf(fp, "environ:\n");
    while (*envp != NULL) {
        fprintf(fp, "%s\n", *envp);
        envp++;
    }
    fclose(fp);
    return 0;
}
EOF