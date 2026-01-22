// SPDX-FileCopyrightText: 2026 The Pion community <https://pion.ly>
// SPDX-License-Identifier: MIT

#include "webrtc.h"
#include <stdio.h>
#include <stdlib.h>
#include <unistd.h>
#include <time.h>
#include <string.h>

void on_data_channel(GoDataChannel d);
void on_open(GoDataChannel d);
void on_message(GoDataChannel d, struct GoDataChannelMessage msg);
char *rand_seq(int n);

int main()
{
    // Register data channel creation handling
    GoRun(on_data_channel);
}

void on_data_channel(GoDataChannel d)
{
    // Register channel opening handling
    GoOnOpen(d, on_open);
    // Register text message handling
    GoOnMessage(d, on_message);
}

void on_open(GoDataChannel d)
{
    char *label = GoLabel(d);
    // d = DataChannel.ID() since we pass this from Go
    printf("Data channel '%s'-'%d' open. Random messages will now be sent to any connected DataChannels every 5 seconds\n", label, d);
    free(label);

    while (1)
    {
        sleep(5);
        char *message = rand_seq(15);
        printf("Sending '%s'\n", message);

        // Send the message as text
        GoSendText(d, message);
        free(message);
    }
}

void on_message(GoDataChannel d, struct GoDataChannelMessage msg)
{
    char *label = GoLabel(d);
    printf("Message from DataChannel '%s': '%s'\n", label, (char *)(msg.data));
    // since we use C.CBytes and C.CString to convert label and msg.data,
    // the converted data are allocated using malloc. So, we need to free them.
    // Reference: https://golang.org/cmd/cgo/#hdr-Go_references_to_C
    free(label);
    free(msg.data);
}

char *rand_seq(int n)
{
    srand(time(0));
    char letters[52] = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ";
    int n_letters = sizeof(letters) / sizeof(char);
    char *b = malloc(sizeof(char) * (n + 1));
    for (int i = 0; i < n; i++)
    {
        b[i] = letters[rand() % n_letters];
    }
    b[n] = '\0';
    return b;
}
